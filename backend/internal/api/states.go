package api

import (
	"database/sql"
	"net/http"

	"techservice/internal/store"
)

func listStates(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		items, err := store.ListStates(r.Context(), database)
		if err != nil {
			writeStoreError(w, err)
			return
		}
		body := make([]map[string]any, 0, len(items))
		for _, item := range items {
			body = append(body, stateJSON(item))
		}
		writeJSON(w, http.StatusOK, body)
	}
}

func createState(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			NameEn     string `json:"nameEn"`
			NameTr     string `json:"nameTr"`
			ShowOnHome bool   `json:"showOnHome"`
		}
		if err := readJSON(w, r, &body); err != nil {
			writeError(w, http.StatusBadRequest, "invalid_json")
			return
		}
		item, err := store.AddState(r.Context(), database, userFrom(r.Context()), store.StateInput{
			NameEn:     body.NameEn,
			NameTr:     body.NameTr,
			ShowOnHome: body.ShowOnHome,
		})
		if err != nil {
			writeStoreError(w, err)
			return
		}
		writeJSON(w, http.StatusCreated, stateJSON(item))
	}
}

func updateState(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			NameEn     *string `json:"nameEn"`
			NameTr     *string `json:"nameTr"`
			Active     *bool   `json:"active"`
			ShowOnHome *bool   `json:"showOnHome"`
		}
		if err := readJSON(w, r, &body); err != nil {
			writeError(w, http.StatusBadRequest, "invalid_json")
			return
		}
		item, err := store.UpdateState(r.Context(), database, userFrom(r.Context()), r.PathValue("code"), store.StatePatch{
			NameEn:     body.NameEn,
			NameTr:     body.NameTr,
			Active:     body.Active,
			ShowOnHome: body.ShowOnHome,
		})
		if err != nil {
			writeStoreError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, stateJSON(item))
	}
}

func moveState(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Position int `json:"position"`
		}
		if err := readJSON(w, r, &body); err != nil {
			writeError(w, http.StatusBadRequest, "invalid_json")
			return
		}
		item, err := store.MoveState(r.Context(), database, userFrom(r.Context()), r.PathValue("code"), body.Position)
		if err != nil {
			writeStoreError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, stateJSON(item))
	}
}

func stateJSON(item store.ServiceState) map[string]any {
	return map[string]any{
		"code":       item.Code,
		"nameEn":     item.NameEn,
		"nameTr":     item.NameTr,
		"position":   item.Position,
		"active":     item.Active,
		"showOnHome": item.ShowOnHome,
		"fixed":      item.Fixed,
		"locked":     item.Locked,
		"used":       item.Used,
	}
}

func manager(database *sql.DB, next http.HandlerFunc) http.HandlerFunc {
	return authenticated(database, func(w http.ResponseWriter, r *http.Request) {
		role := userFrom(r.Context()).Role
		if role != store.RoleAdmin && role != store.RoleOwner {
			writeError(w, http.StatusForbidden, "forbidden")
			return
		}
		next(w, r)
	})
}
