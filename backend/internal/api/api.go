package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"
)

type healthStatus struct {
	Status   string `json:"status"`
	Database string `json:"database"`
}

func New(database *sql.DB) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/health", health(database))
	return mux
}

func health(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		if err := database.PingContext(ctx); err != nil {
			log.Printf("health: database ping: %v", err)
			writeJSON(w, http.StatusServiceUnavailable, healthStatus{Status: "error", Database: "error"})
			return
		}

		writeJSON(w, http.StatusOK, healthStatus{Status: "ok", Database: "ok"})
	}
}

func writeJSON(w http.ResponseWriter, code int, body healthStatus) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(body); err != nil {
		log.Printf("write json: %v", err)
	}
}
