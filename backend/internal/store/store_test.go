package store

import (
	"context"
	"database/sql"
	"errors"
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
