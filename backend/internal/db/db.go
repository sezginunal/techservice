package db

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

func Open(ctx context.Context, path string) (*sql.DB, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("create database directory: %w", err)
	}

	dsn := "file:" + path + "?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)&_pragma=foreign_keys(1)&_txlock=immediate"
	database, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	database.SetMaxOpenConns(4)
	database.SetMaxIdleConns(4)

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var mode string
	if err := database.QueryRowContext(pingCtx, "PRAGMA journal_mode").Scan(&mode); err != nil {
		database.Close()
		return nil, fmt.Errorf("read journal mode: %w", err)
	}
	if mode != "wal" {
		database.Close()
		return nil, fmt.Errorf("journal mode %s", mode)
	}

	if err := database.PingContext(pingCtx); err != nil {
		database.Close()
		return nil, fmt.Errorf("ping sqlite: %w", err)
	}

	return database, nil
}
