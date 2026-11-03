package store

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

const (
	RoleTechnician   = "technician"
	RoleOwner        = "owner"
	RoleAdmin        = "admin"
	sessionLength    = 7 * 24 * time.Hour
	minPassword      = 8
	maxPassword      = 72
	maxNameLen       = 200
	maxNoteLen       = 4000
	maxEmailLen      = 254
	maxPhoneLen      = 40
	maxSerialLen     = 100
	maxPhotoBytes    = 8 << 20
	maxPhotos        = 30
	maxPage          = 100
	defaultPage      = 50
	timestampLayout  = "2006-01-02T15:04:05.000Z"
	StatusRegistered = "registered"
	MaxUploadBody    = int64(maxPhotos)*int64(maxPhotoBytes) + (1 << 20)
	StatusUploadBody = int64(maxPhotoBytes) + (1 << 20)
)

var (
	ErrInvalidCredentials = errors.New("invalid_credentials")
	ErrPasswordTooShort   = errors.New("password_too_short")
	ErrPasswordTooLong    = errors.New("password_too_long")
	ErrUsernameTaken      = errors.New("username_taken")
	ErrInvalidRole        = errors.New("invalid_role")
	ErrLastAdmin          = errors.New("last_admin")
	ErrNotFound           = errors.New("not_found")
	ErrUnauthorized       = errors.New("unauthorized")
	ErrForbidden          = errors.New("forbidden")
	ErrAdminConfig        = errors.New("admin_config")
	ErrInvalidLimit       = errors.New("invalid_limit")
	ErrInvalidSort        = errors.New("invalid_sort")
	ErrInvalidTime        = errors.New("invalid_time")
	ErrInvalidEvent       = errors.New("invalid_event")
	ErrInvalidStatus      = errors.New("invalid_status")

	dummyHash = mustHash("dummy-password")
)

type User struct {
	ID       int64
	Username string
	Role     string
	Active   bool
}

type Session struct {
	ID     string
	User   User
	Expiry time.Time
}

func Migrate(ctx context.Context, database *sql.DB) error {
	_, err := database.ExecContext(ctx, fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY,
			username TEXT NOT NULL UNIQUE COLLATE NOCASE,
			password_hash TEXT NOT NULL,
			role TEXT NOT NULL CHECK (role IN ('technician', 'owner', 'admin')),
			active INTEGER NOT NULL DEFAULT 1 CHECK (active IN (0, 1)),
			created_at TEXT NOT NULL
		);
		CREATE TABLE IF NOT EXISTS sessions (
			id TEXT PRIMARY KEY,
			user_id INTEGER NOT NULL REFERENCES users(id),
			expires_at TEXT NOT NULL,
			created_at TEXT NOT NULL
		);
		CREATE TABLE IF NOT EXISTS customers (
			id INTEGER PRIMARY KEY,
			name TEXT NOT NULL,
			phone TEXT NOT NULL DEFAULT '',
			email TEXT NOT NULL DEFAULT '',
			name_fold TEXT NOT NULL,
			phone_fold TEXT NOT NULL,
			email_fold TEXT NOT NULL,
			created_at TEXT NOT NULL
		);
		CREATE INDEX IF NOT EXISTS customers_name ON customers (name COLLATE NOCASE, id);
		CREATE TABLE IF NOT EXISTS devices (
			id TEXT PRIMARY KEY,
			customer_id INTEGER NOT NULL REFERENCES customers(id),
			brand TEXT NOT NULL,
			model TEXT NOT NULL,
			serial TEXT NOT NULL DEFAULT '',
			note TEXT NOT NULL,
			status TEXT NOT NULL DEFAULT 'registered' %s,
			brand_fold TEXT NOT NULL,
			model_fold TEXT NOT NULL,
			serial_fold TEXT NOT NULL,
			created_at TEXT NOT NULL
		);
		CREATE INDEX IF NOT EXISTS devices_created ON devices (created_at, id);
		CREATE INDEX IF NOT EXISTS devices_customer ON devices (customer_id, created_at, id);
		CREATE INDEX IF NOT EXISTS devices_status ON devices (status, created_at, id);
		CREATE TABLE IF NOT EXISTS device_status_changes (
			id INTEGER PRIMARY KEY,
			device_id TEXT NOT NULL REFERENCES devices(id),
			actor_username TEXT NOT NULL,
			from_status TEXT NOT NULL,
			to_status TEXT NOT NULL,
			note TEXT NOT NULL DEFAULT '',
			photo_path TEXT NOT NULL DEFAULT '',
			photo_size INTEGER NOT NULL DEFAULT 0,
			photo_type TEXT NOT NULL DEFAULT '',
			created_at TEXT NOT NULL
		);
		CREATE INDEX IF NOT EXISTS device_status_changes_device ON device_status_changes (device_id, created_at, id);
		CREATE TABLE IF NOT EXISTS parts (
			id TEXT PRIMARY KEY,
			device_id TEXT NOT NULL REFERENCES devices(id),
			name TEXT NOT NULL,
			note TEXT NOT NULL,
			position INTEGER NOT NULL
		);
		CREATE INDEX IF NOT EXISTS parts_device ON parts (device_id, position);
		CREATE TABLE IF NOT EXISTS photos (
			id TEXT PRIMARY KEY,
			device_id TEXT NOT NULL REFERENCES devices(id),
			part_id TEXT REFERENCES parts(id),
			path TEXT NOT NULL,
			size INTEGER NOT NULL,
			media_type TEXT NOT NULL,
			created_at TEXT NOT NULL
		);
		CREATE INDEX IF NOT EXISTS photos_device ON photos (device_id, created_at, id);
		CREATE TABLE IF NOT EXISTS audit_events (
			id INTEGER PRIMARY KEY,
			created_at TEXT NOT NULL,
			actor_id INTEGER NOT NULL REFERENCES users(id),
			actor_username TEXT NOT NULL,
			actor_username_fold TEXT NOT NULL,
			actor_role TEXT NOT NULL,
			event TEXT NOT NULL,
			subject_fold TEXT NOT NULL,
			detail TEXT NOT NULL
		);
		CREATE INDEX IF NOT EXISTS audit_time ON audit_events (created_at, id);
		CREATE INDEX IF NOT EXISTS audit_role_time ON audit_events (actor_role, created_at, id);
	`, deviceStatusCheck()))
	if err != nil {
		return fmt.Errorf("migrate: %w", err)
	}
	if err := widenDeviceStatus(ctx, database); err != nil {
		return err
	}
	return nil
}

func EnsureAdmin(ctx context.Context, database *sql.DB, username string, password string) error {
	var count int
	if err := database.QueryRowContext(ctx, `SELECT COUNT(*) FROM users WHERE role = ?`, RoleAdmin).Scan(&count); err != nil {
		return fmt.Errorf("count admins: %w", err)
	}
	if count > 0 {
		return nil
	}

	username = strings.TrimSpace(username)
	if username == "" || len(password) < minPassword || len(password) > maxPassword {
		return fmt.Errorf("%w: ADMIN_USERNAME or ADMIN_PASSWORD is missing or invalid", ErrAdminConfig)
	}

	if _, err := insertUser(ctx, database, username, password, RoleAdmin); err != nil {
		return fmt.Errorf("create initial admin: %w", err)
	}
	return nil
}

func Login(ctx context.Context, database *sql.DB, username string, password string) (Session, error) {
	username = strings.TrimSpace(username)
	user, hash, err := findUser(ctx, database, username)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return Session{}, err
	}

	compareHash := dummyHash
	if err == nil {
		compareHash = hash
	}
	passwordErr := bcrypt.CompareHashAndPassword([]byte(compareHash), []byte(password))
	if err != nil || passwordErr != nil || !user.Active {
		return Session{}, ErrInvalidCredentials
	}

	session, err := insertSession(ctx, database, user.ID)
	if err != nil {
		return Session{}, err
	}
	session.User = user
	return session, nil
}

func UserFromSession(ctx context.Context, database *sql.DB, sessionID string) (User, error) {
	if sessionID == "" {
		return User{}, ErrUnauthorized
	}

	var user User
	var active int
	var expiresAt string
	err := database.QueryRowContext(ctx, `
		SELECT users.id, users.username, users.role, users.active, sessions.expires_at
		FROM sessions
		JOIN users ON users.id = sessions.user_id
		WHERE sessions.id = ?
	`, sessionID).Scan(&user.ID, &user.Username, &user.Role, &active, &expiresAt)
	if errors.Is(err, sql.ErrNoRows) {
		return User{}, ErrUnauthorized
	}
	if err != nil {
		return User{}, fmt.Errorf("read session: %w", err)
	}

	expiry, err := time.Parse(time.RFC3339Nano, expiresAt)
	if err != nil || !expiry.After(time.Now().UTC()) || active == 0 {
		if _, deleteErr := database.ExecContext(ctx, `DELETE FROM sessions WHERE id = ?`, sessionID); deleteErr != nil {
			return User{}, fmt.Errorf("delete expired session: %w", deleteErr)
		}
		return User{}, ErrUnauthorized
	}

	user.Active = true
	return user, nil
}

func Logout(ctx context.Context, database *sql.DB, sessionID string) error {
	if sessionID == "" {
		return nil
	}
	if _, err := database.ExecContext(ctx, `DELETE FROM sessions WHERE id = ?`, sessionID); err != nil {
		return fmt.Errorf("delete session: %w", err)
	}
	return nil
}

func ListUsers(ctx context.Context, database *sql.DB) ([]User, error) {
	rows, err := database.QueryContext(ctx, `
		SELECT id, username, role, active
		FROM users
		ORDER BY username COLLATE NOCASE
	`)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		var active int
		if err := rows.Scan(&user.ID, &user.Username, &user.Role, &active); err != nil {
			return nil, fmt.Errorf("scan user: %w", err)
		}
		user.Active = active == 1
		users = append(users, user)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	return users, nil
}

func CreateUser(ctx context.Context, database *sql.DB, actor User, username string, password string, role string) (User, error) {
	username = strings.TrimSpace(username)
	if username == "" || len(username) > 64 {
		return User{}, ErrInvalidCredentials
	}
	if !validRole(role) {
		return User{}, ErrInvalidRole
	}
	if err := validatePassword(password); err != nil {
		return User{}, err
	}

	tx, err := database.BeginTx(ctx, nil)
	if err != nil {
		return User{}, fmt.Errorf("begin create user: %w", err)
	}
	defer tx.Rollback()

	id, err := insertUser(ctx, tx, username, password, role)
	if err != nil {
		return User{}, err
	}
	if err := insertAudit(ctx, tx, actor, EventUserCreated, username+" "+role, map[string]string{
		"username": username,
		"role":     role,
	}); err != nil {
		return User{}, err
	}
	if err := tx.Commit(); err != nil {
		return User{}, fmt.Errorf("commit create user: %w", err)
	}
	return User{ID: id, Username: username, Role: role, Active: true}, nil
}

type UserPatch struct {
	Role     *string
	Password *string
	Active   *bool
}

func UpdateUser(ctx context.Context, database *sql.DB, actor User, id int64, patch UserPatch) (User, error) {
	tx, err := database.BeginTx(ctx, nil)
	if err != nil {
		return User{}, fmt.Errorf("begin update: %w", err)
	}
	defer tx.Rollback()

	var user User
	var active int
	var hash string
	err = tx.QueryRowContext(ctx, `
		SELECT id, username, role, active, password_hash
		FROM users
		WHERE id = ?
	`, id).Scan(&user.ID, &user.Username, &user.Role, &active, &hash)
	if errors.Is(err, sql.ErrNoRows) {
		return User{}, ErrNotFound
	}
	if err != nil {
		return User{}, fmt.Errorf("read user: %w", err)
	}
	user.Active = active == 1

	nextRole := user.Role
	if patch.Role != nil {
		if !validRole(*patch.Role) {
			return User{}, ErrInvalidRole
		}
		nextRole = *patch.Role
	}
	nextActive := user.Active
	if patch.Active != nil {
		nextActive = *patch.Active
	}

	if user.Role == RoleAdmin && user.Active && (nextRole != RoleAdmin || !nextActive) {
		var others int
		if err := tx.QueryRowContext(ctx, `
			SELECT COUNT(*) FROM users
			WHERE role = ? AND active = 1 AND id != ?
		`, RoleAdmin, id).Scan(&others); err != nil {
			return User{}, fmt.Errorf("count other admins: %w", err)
		}
		if others == 0 {
			return User{}, ErrLastAdmin
		}
	}

	passwordChanged := false
	if patch.Password != nil && *patch.Password != "" {
		if err := validatePassword(*patch.Password); err != nil {
			return User{}, err
		}
		hashed, err := bcrypt.GenerateFromPassword([]byte(*patch.Password), bcrypt.DefaultCost)
		if err != nil {
			return User{}, fmt.Errorf("hash password: %w", err)
		}
		hash = string(hashed)
		passwordChanged = true
	}

	activeValue := 0
	if nextActive {
		activeValue = 1
	}
	if _, err := tx.ExecContext(ctx, `
		UPDATE users
		SET role = ?, active = ?, password_hash = ?
		WHERE id = ?
	`, nextRole, activeValue, hash, id); err != nil {
		return User{}, fmt.Errorf("update user: %w", err)
	}
	if passwordChanged || !nextActive {
		if _, err := tx.ExecContext(ctx, `DELETE FROM sessions WHERE user_id = ?`, id); err != nil {
			return User{}, fmt.Errorf("delete user sessions: %w", err)
		}
	}
	if nextRole != user.Role {
		if err := insertAudit(ctx, tx, actor, EventUserRoleChanged, user.Username+" "+user.Role+" "+nextRole, map[string]string{
			"username": user.Username,
			"from":     user.Role,
			"to":       nextRole,
		}); err != nil {
			return User{}, err
		}
	}
	if passwordChanged {
		if err := insertAudit(ctx, tx, actor, EventUserPasswordChanged, user.Username, map[string]string{
			"username": user.Username,
		}); err != nil {
			return User{}, err
		}
	}
	if nextActive != user.Active {
		event := EventUserDeactivated
		if nextActive {
			event = EventUserActivated
		}
		if err := insertAudit(ctx, tx, actor, event, user.Username, map[string]string{
			"username": user.Username,
		}); err != nil {
			return User{}, err
		}
	}
	if err := tx.Commit(); err != nil {
		return User{}, fmt.Errorf("commit update: %w", err)
	}

	user.Role = nextRole
	user.Active = nextActive
	return user, nil
}

type queryer interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
	QueryRowContext(context.Context, string, ...any) *sql.Row
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
}

func insertUser(ctx context.Context, database queryer, username string, password string, role string) (int64, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return 0, fmt.Errorf("hash password: %w", err)
	}
	result, err := database.ExecContext(ctx, `
		INSERT INTO users (username, password_hash, role, active, created_at)
		VALUES (?, ?, ?, 1, ?)
	`, username, string(hashed), role, time.Now().UTC().Format(time.RFC3339Nano))
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			return 0, ErrUsernameTaken
		}
		return 0, fmt.Errorf("insert user: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("user id: %w", err)
	}
	return id, nil
}

func insertSession(ctx context.Context, database *sql.DB, userID int64) (Session, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return Session{}, fmt.Errorf("session id: %w", err)
	}
	now := time.Now().UTC()
	session := Session{
		ID:     hex.EncodeToString(buf),
		Expiry: now.Add(sessionLength),
	}
	_, err := database.ExecContext(ctx, `
		INSERT INTO sessions (id, user_id, expires_at, created_at)
		VALUES (?, ?, ?, ?)
	`, session.ID, userID, session.Expiry.Format(time.RFC3339Nano), now.Format(time.RFC3339Nano))
	if err != nil {
		return Session{}, fmt.Errorf("insert session: %w", err)
	}
	return session, nil
}

func findUser(ctx context.Context, database *sql.DB, username string) (User, string, error) {
	var user User
	var active int
	var hash string
	err := database.QueryRowContext(ctx, `
		SELECT id, username, role, active, password_hash
		FROM users
		WHERE username = ?
	`, username).Scan(&user.ID, &user.Username, &user.Role, &active, &hash)
	user.Active = active == 1
	return user, hash, err
}

func validatePassword(password string) error {
	if len(password) < minPassword {
		return ErrPasswordTooShort
	}
	if len(password) > maxPassword {
		return ErrPasswordTooLong
	}
	return nil
}

func validRole(role string) bool {
	return role == RoleTechnician || role == RoleOwner || role == RoleAdmin
}

func mustHash(password string) string {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}
	return string(hashed)
}
