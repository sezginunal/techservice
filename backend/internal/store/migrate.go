package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

func Migrate(ctx context.Context, database *sql.DB) error {
	if _, err := database.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version INTEGER PRIMARY KEY,
			applied_at TEXT NOT NULL
		)
	`); err != nil {
		return fmt.Errorf("migrate: %w", err)
	}
	var version int
	if err := database.QueryRowContext(ctx, `SELECT COALESCE(MAX(version), 0) FROM schema_migrations`).Scan(&version); err != nil {
		return fmt.Errorf("migrate version: %w", err)
	}
	if version == 0 {
		var ready int
		if err := database.QueryRowContext(ctx, `SELECT COUNT(*) FROM sqlite_master WHERE type = 'table' AND name = 'users'`).Scan(&ready); err != nil {
			return fmt.Errorf("migrate baseline: %w", err)
		}
		if ready == 1 {
			if err := recordMigration(ctx, database, 1); err != nil {
				return err
			}
			version = 1
		}
	}
	steps := []struct {
		version int
		fkOff   bool
		run     func(context.Context, *sql.Tx) error
	}{
		{1, false, migration1},
		{2, true, migration2},
		{3, false, migration3},
	}
	for _, step := range steps {
		if version >= step.version {
			continue
		}
		if err := applyMigration(ctx, database, step.version, step.fkOff, step.run); err != nil {
			return err
		}
		version = step.version
	}
	return nil
}

func applyMigration(ctx context.Context, database *sql.DB, version int, fkOff bool, run func(context.Context, *sql.Tx) error) error {
	conn, err := database.Conn(ctx)
	if err != nil {
		return fmt.Errorf("migrate %d: %w", version, err)
	}
	defer conn.Close()
	if fkOff {
		if _, err := conn.ExecContext(ctx, `PRAGMA foreign_keys = OFF`); err != nil {
			return fmt.Errorf("migrate %d: %w", version, err)
		}
		defer conn.ExecContext(ctx, `PRAGMA foreign_keys = ON`)
	}
	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("migrate %d: %w", version, err)
	}
	defer tx.Rollback()
	if err := run(ctx, tx); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO schema_migrations (version, applied_at) VALUES (?, ?)`, version, timestamp(time.Now())); err != nil {
		return fmt.Errorf("migrate %d: %w", version, err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("migrate %d: %w", version, err)
	}
	if fkOff {
		rows, err := conn.QueryContext(ctx, `PRAGMA foreign_key_check`)
		if err != nil {
			return fmt.Errorf("migrate %d check: %w", version, err)
		}
		defer rows.Close()
		if rows.Next() {
			return fmt.Errorf("migrate %d: foreign key check failed", version)
		}
		if err := rows.Err(); err != nil {
			return fmt.Errorf("migrate %d check: %w", version, err)
		}
	}
	return nil
}

func recordMigration(ctx context.Context, database *sql.DB, version int) error {
	if _, err := database.ExecContext(ctx, `INSERT INTO schema_migrations (version, applied_at) VALUES (?, ?)`, version, timestamp(time.Now())); err != nil {
		return fmt.Errorf("migrate %d: %w", version, err)
	}
	return nil
}

func migration1(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.ExecContext(ctx, fmt.Sprintf(`
		CREATE TABLE users (
			id INTEGER PRIMARY KEY,
			username TEXT NOT NULL UNIQUE COLLATE NOCASE,
			password_hash TEXT NOT NULL,
			role TEXT NOT NULL CHECK (role IN ('technician', 'owner', 'admin')),
			active INTEGER NOT NULL DEFAULT 1 CHECK (active IN (0, 1)),
			created_at TEXT NOT NULL
		);
		CREATE TABLE sessions (
			id TEXT PRIMARY KEY,
			user_id INTEGER NOT NULL REFERENCES users(id),
			expires_at TEXT NOT NULL,
			created_at TEXT NOT NULL
		);
		CREATE TABLE customers (
			id INTEGER PRIMARY KEY,
			name TEXT NOT NULL,
			phone TEXT NOT NULL,
			email TEXT NOT NULL DEFAULT '',
			address TEXT NOT NULL DEFAULT '',
			name_fold TEXT NOT NULL,
			phone_fold TEXT NOT NULL,
			email_fold TEXT NOT NULL,
			address_fold TEXT NOT NULL DEFAULT '',
			created_at TEXT NOT NULL
		);
		CREATE INDEX customers_name ON customers (name COLLATE NOCASE, id);
		CREATE UNIQUE INDEX customers_phone ON customers (phone);
		CREATE TABLE devices (
			id TEXT PRIMARY KEY,
			brand TEXT NOT NULL,
			model TEXT NOT NULL,
			serial TEXT NOT NULL,
			brand_fold TEXT NOT NULL,
			model_fold TEXT NOT NULL,
			serial_fold TEXT NOT NULL,
			created_at TEXT NOT NULL
		);
		CREATE UNIQUE INDEX devices_identity ON devices (brand_fold, model_fold, serial_fold);
		CREATE TABLE services (
			id TEXT PRIMARY KEY,
			device_id TEXT NOT NULL REFERENCES devices(id),
			customer_id INTEGER NOT NULL REFERENCES customers(id),
			note TEXT NOT NULL,
			status TEXT NOT NULL DEFAULT 'registered' %s,
			created_at TEXT NOT NULL
		);
		CREATE INDEX services_created ON services (created_at, id);
		CREATE INDEX services_customer ON services (customer_id, created_at, id);
		CREATE INDEX services_device ON services (device_id, created_at, id);
		CREATE INDEX services_status ON services (status, created_at, id);
		CREATE UNIQUE INDEX services_one_open ON services (device_id) WHERE status NOT IN ('delivered', 'cancelled');
		CREATE TABLE service_status_changes (
			id INTEGER PRIMARY KEY,
			service_id TEXT NOT NULL REFERENCES services(id),
			actor_username TEXT NOT NULL,
			from_status TEXT NOT NULL,
			to_status TEXT NOT NULL,
			note TEXT NOT NULL DEFAULT '',
			photo_path TEXT NOT NULL DEFAULT '',
			photo_size INTEGER NOT NULL DEFAULT 0,
			photo_type TEXT NOT NULL DEFAULT '',
			created_at TEXT NOT NULL
		);
		CREATE INDEX service_status_changes_service ON service_status_changes (service_id, created_at, id);
		CREATE TABLE parts (
			id TEXT PRIMARY KEY,
			service_id TEXT NOT NULL REFERENCES services(id),
			name TEXT NOT NULL,
			note TEXT NOT NULL,
			position INTEGER NOT NULL
		);
		CREATE INDEX parts_service ON parts (service_id, position);
		CREATE TABLE photos (
			id TEXT PRIMARY KEY,
			service_id TEXT NOT NULL REFERENCES services(id),
			part_id TEXT REFERENCES parts(id),
			path TEXT NOT NULL,
			size INTEGER NOT NULL,
			media_type TEXT NOT NULL,
			created_at TEXT NOT NULL
		);
		CREATE INDEX photos_service ON photos (service_id, created_at, id);
		CREATE TABLE audit_events (
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
		CREATE INDEX audit_time ON audit_events (created_at, id);
		CREATE INDEX audit_role_time ON audit_events (actor_role, created_at, id);
	`, deviceStatusCheck()))
	if err != nil {
		return fmt.Errorf("migrate 1: %w", err)
	}
	return nil
}

func migration2(ctx context.Context, tx *sql.Tx) error {
	if _, err := tx.ExecContext(ctx, `
		CREATE TABLE service_states (
			code TEXT PRIMARY KEY,
			name_en TEXT NOT NULL,
			name_tr TEXT NOT NULL,
			name_en_fold TEXT NOT NULL UNIQUE,
			name_tr_fold TEXT NOT NULL UNIQUE,
			position INTEGER NOT NULL UNIQUE,
			active INTEGER NOT NULL CHECK (active IN (0, 1)),
			show_on_home INTEGER NOT NULL CHECK (show_on_home IN (0, 1))
		)
	`); err != nil {
		return fmt.Errorf("migrate 2: %w", err)
	}
	if err := seedServiceStates(ctx, tx); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `
		CREATE TABLE services_next (
			id TEXT PRIMARY KEY,
			device_id TEXT NOT NULL REFERENCES devices(id),
			customer_id INTEGER NOT NULL REFERENCES customers(id),
			note TEXT NOT NULL,
			status TEXT NOT NULL DEFAULT 'registered',
			created_at TEXT NOT NULL
		);
		INSERT INTO services_next (id, device_id, customer_id, note, status, created_at)
		SELECT id, device_id, customer_id, note, status, created_at FROM services;
		DROP TABLE services;
		ALTER TABLE services_next RENAME TO services;
		CREATE INDEX services_created ON services (created_at, id);
		CREATE INDEX services_customer ON services (customer_id, created_at, id);
		CREATE INDEX services_device ON services (device_id, created_at, id);
		CREATE INDEX services_status ON services (status, created_at, id);
		CREATE UNIQUE INDEX services_one_open ON services (device_id) WHERE status NOT IN ('delivered', 'cancelled');
	`); err != nil {
		return fmt.Errorf("migrate 2: %w", err)
	}
	return nil
}

func migration3(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.ExecContext(ctx, `
		CREATE TABLE spare_parts (
			id TEXT PRIMARY KEY,
			brand TEXT NOT NULL,
			model TEXT NOT NULL,
			brand_fold TEXT NOT NULL,
			model_fold TEXT NOT NULL,
			on_hand INTEGER NOT NULL CHECK (on_hand >= 0 AND on_hand <= 99999),
			created_at TEXT NOT NULL
		);
		CREATE UNIQUE INDEX spare_parts_identity ON spare_parts (brand_fold, model_fold);
		CREATE INDEX spare_parts_order ON spare_parts (brand COLLATE NOCASE, model COLLATE NOCASE, id);
		CREATE TABLE spare_part_moves (
			id INTEGER PRIMARY KEY,
			part_id TEXT NOT NULL REFERENCES spare_parts(id),
			kind TEXT NOT NULL CHECK (kind IN ('receipt', 'write_off', 'use', 'return')),
			quantity INTEGER NOT NULL CHECK (quantity >= 1 AND quantity <= 99999),
			note TEXT NOT NULL DEFAULT '',
			service_id TEXT REFERENCES services(id),
			state_code TEXT,
			actor_username TEXT NOT NULL,
			created_at TEXT NOT NULL,
			CHECK (
				(kind IN ('receipt', 'write_off') AND service_id IS NULL AND state_code IS NULL)
				OR (kind IN ('use', 'return') AND service_id IS NOT NULL AND state_code IS NOT NULL AND state_code <> '')
			)
		);
		CREATE INDEX spare_part_moves_stock ON spare_part_moves (part_id, created_at, id) WHERE kind IN ('receipt', 'write_off');
		CREATE INDEX spare_part_moves_use ON spare_part_moves (part_id, created_at, id) WHERE kind IN ('use', 'return');
		CREATE INDEX spare_part_moves_service_use ON spare_part_moves (service_id, created_at, id) WHERE kind IN ('use', 'return');
		CREATE INDEX spare_part_moves_service_part ON spare_part_moves (service_id, part_id) WHERE kind IN ('use', 'return');
	`)
	if err != nil {
		return fmt.Errorf("migrate 3: %w", err)
	}
	return nil
}
