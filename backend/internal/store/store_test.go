package store

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"
	"time"

	"techservice/internal/db"
)

func TestLoginAndSessionExpiry(t *testing.T) {
	ctx := context.Background()
	database := openTestDB(t)

	if err := EnsureAdmin(ctx, database, "admin", "admin1234"); err != nil {
		t.Fatalf("ensure admin: %v", err)
	}
	if err := EnsureAdmin(ctx, database, "other", "other-password"); err != nil {
		t.Fatalf("second ensure: %v", err)
	}

	var count int
	if err := database.QueryRowContext(ctx, `SELECT COUNT(*) FROM users`).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("admin count %d", count)
	}

	session, err := Login(ctx, database, "admin", "admin1234")
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	if _, err := Login(ctx, database, "admin", "wrong-password"); !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("bad password: %v", err)
	}

	past := time.Now().UTC().Add(-time.Hour).Format(time.RFC3339Nano)
	if _, err := database.ExecContext(ctx, `UPDATE sessions SET expires_at = ? WHERE id = ?`, past, session.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := UserFromSession(ctx, database, session.ID); !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("expired session: %v", err)
	}
}

func TestPasswordAndLastAdmin(t *testing.T) {
	ctx := context.Background()
	database := openTestDB(t)
	if err := EnsureAdmin(ctx, database, "admin", "admin1234"); err != nil {
		t.Fatal(err)
	}
	admin, err := Login(ctx, database, "admin", "admin1234")
	if err != nil {
		t.Fatal(err)
	}

	if _, err := CreateUser(ctx, database, admin.User, "tech", "short", RoleTechnician); !errors.Is(err, ErrPasswordTooShort) {
		t.Fatalf("short password: %v", err)
	}
	if _, err := CreateUser(ctx, database, admin.User, "tech", "technician1", RoleTechnician); err != nil {
		t.Fatal(err)
	}

	role := RoleTechnician
	if _, err := UpdateUser(ctx, database, admin.User, admin.User.ID, UserPatch{Role: &role}); !errors.Is(err, ErrLastAdmin) {
		t.Fatalf("last admin: %v", err)
	}
}

func TestTurkishUsernameFold(t *testing.T) {
	ctx := context.Background()
	database := openTestDB(t)
	if err := EnsureAdmin(ctx, database, "admin", "admin1234"); err != nil {
		t.Fatal(err)
	}
	admin, err := Login(ctx, database, "ADMIN", "admin1234")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := CreateUser(ctx, database, admin.User, "Şahin", "technician1", RoleTechnician); err != nil {
		t.Fatal(err)
	}
	if _, err := CreateUser(ctx, database, admin.User, "sahin", "technician1", RoleTechnician); !errors.Is(err, ErrUsernameTaken) {
		t.Fatalf("folded username: %v", err)
	}
	if _, err := CreateUser(ctx, database, admin.User, "İsa", "technician1", RoleTechnician); err != nil {
		t.Fatal(err)
	}
	if _, err := CreateUser(ctx, database, admin.User, "isa", "technician1", RoleTechnician); !errors.Is(err, ErrUsernameTaken) {
		t.Fatalf("dotted i username: %v", err)
	}
	session, err := Login(ctx, database, "şahin", "technician1")
	if err != nil {
		t.Fatalf("login folded username: %v", err)
	}
	if session.User.Username != "Şahin" {
		t.Fatalf("stored username %q", session.User.Username)
	}
	if _, err := CreateUser(ctx, database, admin.User, strings.Repeat("ş", 33), "technician1", RoleTechnician); err != nil {
		t.Fatalf("33-letter username: %v", err)
	}
	if _, err := CreateUser(ctx, database, admin.User, strings.Repeat("ş", 65), "technician1", RoleTechnician); !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("65-letter username: %v", err)
	}
	if _, err := CreateUser(ctx, database, admin.User, "euro", strings.Repeat("€", 40), RoleTechnician); !errors.Is(err, ErrPasswordTooLong) {
		t.Fatalf("long password bytes: %v", err)
	}
	if _, err := CreateUser(ctx, database, admin.User, "okpass", strings.Repeat("a", 72), RoleTechnician); err != nil {
		t.Fatalf("72 ascii password: %v", err)
	}
}

func openTestDB(t *testing.T) *sql.DB {
	t.Helper()
	database, err := db.Open(context.Background(), t.TempDir()+"/test.db")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { database.Close() })
	if err := Migrate(context.Background(), database); err != nil {
		t.Fatal(err)
	}
	return database
}
