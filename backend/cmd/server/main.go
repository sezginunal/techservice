package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"techservice/internal/api"
	"techservice/internal/db"
	"techservice/internal/store"
)

func main() {
	log.SetFlags(log.LstdFlags | log.LUTC)

	if len(os.Args) > 1 && os.Args[1] == "-healthcheck" {
		os.Exit(runHealthcheck())
	}

	databasePath := env("DATABASE_PATH", "/data/techservice.db")
	uploadPath := env("UPLOAD_PATH", "/uploads")
	addr := env("ADDR", ":8080")

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	database, err := db.Open(ctx, databasePath)
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	defer database.Close()

	if err := os.MkdirAll(uploadPath, 0o755); err != nil {
		log.Fatalf("create upload directory: %v", err)
	}

	if err := store.Migrate(ctx, database); err != nil {
		log.Fatalf("migrate database: %v", err)
	}
	if err := store.EnsureAdmin(ctx, database, os.Getenv("ADMIN_USERNAME"), os.Getenv("ADMIN_PASSWORD")); err != nil {
		log.Fatalf("initial admin: %v", err)
	}

	log.Printf("database %s", databasePath)
	log.Printf("uploads %s", uploadPath)

	srv := &http.Server{
		Addr:              addr,
		Handler:           api.New(database),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		log.Printf("listening on %s", addr)
		errCh <- srv.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			log.Printf("shutdown: %v", err)
		}
	case err := <-errCh:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("listen: %v", err)
		}
	}
}

func runHealthcheck() int {
	client := &http.Client{Timeout: 2 * time.Second}
	response, err := client.Get(healthcheckURL())
	if err != nil {
		log.Printf("healthcheck: %v", err)
		return 1
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		log.Printf("healthcheck: status %d", response.StatusCode)
		return 1
	}
	return 0
}

func healthcheckURL() string {
	addr := env("ADDR", ":8080")
	if strings.HasPrefix(addr, ":") {
		return "http://127.0.0.1" + addr + "/api/health"
	}
	return "http://" + addr + "/api/health"
}

func env(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
