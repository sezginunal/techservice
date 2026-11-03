package api

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"

	"techservice/internal/brand"
	"techservice/internal/store"
)

const sessionCookie = "session"

type userResponse struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	Active   bool   `json:"active"`
}

func New(database *sql.DB, uploadPath string, shop brand.Brand) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/health", health(database))
	mux.HandleFunc("GET /api/brand", showBrand(shop))
	mux.HandleFunc("GET /api/brand/logo", showLogo(shop))
	mux.HandleFunc("POST /api/login", login(database))
	mux.HandleFunc("POST /api/logout", authenticated(database, logout(database)))
	mux.HandleFunc("GET /api/session", authenticated(database, currentSession()))
	mux.HandleFunc("GET /api/users", admin(database, listUsers(database)))
	mux.HandleFunc("POST /api/users", admin(database, createUser(database)))
	mux.HandleFunc("PATCH /api/users/{id}", admin(database, updateUser(database)))
	mux.HandleFunc("GET /api/customers", authenticated(database, listCustomers(database)))
	mux.HandleFunc("GET /api/customers/{id}", authenticated(database, getCustomer(database)))
	mux.HandleFunc("PATCH /api/customers/{id}", authenticated(database, updateCustomer(database)))
	mux.HandleFunc("GET /api/devices", authenticated(database, listDevices(database)))
	mux.HandleFunc("POST /api/devices", authenticated(database, registerDevice(database, uploadPath)))
	mux.HandleFunc("GET /api/devices/{id}/photos/{photoId}", authenticated(database, getPhoto(database, uploadPath)))
	mux.HandleFunc("GET /api/devices/{id}/status/{changeId}/photo", authenticated(database, getStatusPhoto(database, uploadPath)))
	mux.HandleFunc("POST /api/devices/{id}/status", authenticated(database, changeDeviceStatus(database, uploadPath)))
	mux.HandleFunc("GET /api/devices/{id}", authenticated(database, getDevice(database)))
	mux.HandleFunc("PUT /api/devices/{id}", authenticated(database, updateDevice(database, uploadPath)))
	mux.HandleFunc("GET /api/audit", authenticated(database, listAudit(database)))
	return mux
}

func showBrand(shop brand.Brand) http.HandlerFunc {
	_, _, logo := shop.Logo()
	body := map[string]any{"name": shop.Name, "logo": logo}
	return func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, body)
	}
}

func showLogo(shop brand.Brand) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, media, ok := shop.Logo()
		if !ok {
			writeError(w, http.StatusNotFound, "not_found")
			return
		}
		w.Header().Set("Content-Type", media)
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Content-Security-Policy", "default-src 'none'; sandbox")
		w.Header().Set("Cache-Control", "no-cache")
		http.ServeContent(w, r, "", time.Time{}, bytes.NewReader(body))
	}
}

func health(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		if err := database.PingContext(ctx); err != nil {
			log.Printf("health: database ping: %v", err)
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "error", "database": "error"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "database": "ok"})
	}
}

func login(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}
		if err := readJSON(w, r, &body); err != nil {
			writeError(w, http.StatusBadRequest, "invalid_json")
			return
		}

		session, err := store.Login(r.Context(), database, body.Username, body.Password)
		if errors.Is(err, store.ErrInvalidCredentials) {
			log.Printf("login: rejected")
			writeError(w, http.StatusUnauthorized, "invalid_credentials")
			return
		}
		if err != nil {
			log.Printf("login: %v", err)
			writeError(w, http.StatusInternalServerError, "internal")
			return
		}

		setSessionCookie(w, session)
		writeJSON(w, http.StatusOK, userResponseFrom(session.User))
	}
}

func logout(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sessionID := cookieValue(r)
		if err := store.Logout(r.Context(), database, sessionID); err != nil {
			log.Printf("logout: %v", err)
			writeError(w, http.StatusInternalServerError, "internal")
			return
		}
		clearSessionCookie(w)
		w.WriteHeader(http.StatusNoContent)
	}
}

func currentSession() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, userResponseFrom(userFrom(r.Context())))
	}
}

func listUsers(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		users, err := store.ListUsers(r.Context(), database)
		if err != nil {
			log.Printf("list users: %v", err)
			writeError(w, http.StatusInternalServerError, "internal")
			return
		}
		body := make([]userResponse, 0, len(users))
		for _, user := range users {
			body = append(body, userResponseFrom(user))
		}
		writeJSON(w, http.StatusOK, body)
	}
}

func createUser(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Username string `json:"username"`
			Password string `json:"password"`
			Role     string `json:"role"`
		}
		if err := readJSON(w, r, &body); err != nil {
			writeError(w, http.StatusBadRequest, "invalid_json")
			return
		}

		user, err := store.CreateUser(r.Context(), database, userFrom(r.Context()), body.Username, body.Password, body.Role)
		if err != nil {
			writeStoreError(w, err)
			return
		}
		writeJSON(w, http.StatusCreated, userResponseFrom(user))
	}
}

func updateUser(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid_id")
			return
		}

		var body struct {
			Role     *string `json:"role"`
			Password *string `json:"password"`
			Active   *bool   `json:"active"`
		}
		if err := readJSON(w, r, &body); err != nil {
			writeError(w, http.StatusBadRequest, "invalid_json")
			return
		}

		user, err := store.UpdateUser(r.Context(), database, userFrom(r.Context()), id, store.UserPatch{
			Role:     body.Role,
			Password: body.Password,
			Active:   body.Active,
		})
		if err != nil {
			writeStoreError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, userResponseFrom(user))
	}
}

func authenticated(database *sql.DB, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, err := store.UserFromSession(r.Context(), database, cookieValue(r))
		if errors.Is(err, store.ErrUnauthorized) {
			clearSessionCookie(w)
			writeError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		if err != nil {
			log.Printf("session: %v", err)
			writeError(w, http.StatusInternalServerError, "internal")
			return
		}
		next(w, r.WithContext(context.WithValue(r.Context(), userKey, user)))
	}
}

func admin(database *sql.DB, next http.HandlerFunc) http.HandlerFunc {
	return authenticated(database, func(w http.ResponseWriter, r *http.Request) {
		if userFrom(r.Context()).Role != store.RoleAdmin {
			writeError(w, http.StatusForbidden, "forbidden")
			return
		}
		next(w, r)
	})
}

type contextKey int

const userKey contextKey = 1

func userFrom(ctx context.Context) store.User {
	user, _ := ctx.Value(userKey).(store.User)
	return user
}

func cookieValue(r *http.Request) string {
	cookie, err := r.Cookie(sessionCookie)
	if err != nil {
		return ""
	}
	return cookie.Value
}

func setSessionCookie(w http.ResponseWriter, session store.Session) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookie,
		Value:    session.ID,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Expires:  session.Expiry,
		MaxAge:   int(time.Until(session.Expiry).Seconds()),
	})
}

func clearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookie,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})
}

func readJSON(w http.ResponseWriter, r *http.Request, dest any) error {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	return decoder.Decode(dest)
}

func writeStoreError(w http.ResponseWriter, err error) {
	var fields *store.FieldError
	if errors.As(err, &fields) {
		log.Printf("validation: %v", err)
		writeValidation(w, fields)
		return
	}
	switch {
	case errors.Is(err, store.ErrPasswordTooShort):
		writeError(w, http.StatusBadRequest, "password_too_short")
	case errors.Is(err, store.ErrPasswordTooLong):
		writeError(w, http.StatusBadRequest, "password_too_long")
	case errors.Is(err, store.ErrUsernameTaken):
		writeError(w, http.StatusConflict, "username_taken")
	case errors.Is(err, store.ErrInvalidRole):
		writeError(w, http.StatusBadRequest, "invalid_role")
	case errors.Is(err, store.ErrLastAdmin):
		writeError(w, http.StatusConflict, "last_admin")
	case errors.Is(err, store.ErrNotFound):
		writeError(w, http.StatusNotFound, "not_found")
	case errors.Is(err, store.ErrInvalidCredentials):
		writeError(w, http.StatusBadRequest, "invalid_username")
	case errors.Is(err, store.ErrForbidden):
		writeError(w, http.StatusForbidden, "forbidden")
	case errors.Is(err, store.ErrInvalidLimit):
		writeError(w, http.StatusBadRequest, "invalid_limit")
	case errors.Is(err, store.ErrInvalidSort):
		writeError(w, http.StatusBadRequest, "invalid_sort")
	case errors.Is(err, store.ErrInvalidTime):
		writeError(w, http.StatusBadRequest, "invalid_time")
	case errors.Is(err, store.ErrInvalidEvent):
		writeError(w, http.StatusBadRequest, "invalid_event")
	case errors.Is(err, store.ErrInvalidStatus):
		writeError(w, http.StatusBadRequest, "invalid_status")
	default:
		log.Printf("store: %v", err)
		writeError(w, http.StatusInternalServerError, "internal")
	}
}

func writeValidation(w http.ResponseWriter, fields *store.FieldError) {
	missing := fields.Missing
	invalid := fields.Invalid
	tooLong := fields.TooLong
	if missing == nil {
		missing = []string{}
	}
	if invalid == nil {
		invalid = []string{}
	}
	if tooLong == nil {
		tooLong = []string{}
	}
	writeJSON(w, http.StatusBadRequest, map[string]any{
		"error":   "validation",
		"missing": missing,
		"invalid": invalid,
		"tooLong": tooLong,
		"photo":   fields.Photo,
	})
}

func writeError(w http.ResponseWriter, code int, message string) {
	writeJSON(w, code, map[string]string{"error": message})
}

func writeJSON(w http.ResponseWriter, code int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(body); err != nil {
		log.Printf("write json: %v", err)
	}
}

func userResponseFrom(user store.User) userResponse {
	return userResponse{
		ID:       user.ID,
		Username: user.Username,
		Role:     user.Role,
		Active:   user.Active,
	}
}
