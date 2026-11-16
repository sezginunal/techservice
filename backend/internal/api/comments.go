package api

import (
	"database/sql"
	"net/http"
	"strconv"

	"techservice/internal/store"
)

func addServiceComment(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Text string `json:"text"`
		}
		if err := readJSON(w, r, &body); err != nil {
			writeError(w, http.StatusBadRequest, "invalid_json")
			return
		}
		comment, err := store.AddServiceComment(r.Context(), database, userFrom(r.Context()), r.PathValue("id"), body.Text)
		if err != nil {
			writeStoreError(w, err)
			return
		}
		writeJSON(w, http.StatusCreated, commentJSON(comment))
	}
}

func updateServiceComment(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		commentID, err := strconv.ParseInt(r.PathValue("commentId"), 10, 64)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid_id")
			return
		}
		var body struct {
			Text string `json:"text"`
		}
		if err := readJSON(w, r, &body); err != nil {
			writeError(w, http.StatusBadRequest, "invalid_json")
			return
		}
		comment, err := store.UpdateServiceComment(r.Context(), database, userFrom(r.Context()), r.PathValue("id"), commentID, body.Text)
		if err != nil {
			writeStoreError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, commentJSON(comment))
	}
}

func commentJSON(comment store.ServiceComment) map[string]any {
	return map[string]any{
		"id":             comment.ID,
		"authorId":       comment.AuthorID,
		"authorUsername": comment.AuthorUsername,
		"text":           comment.Body,
		"createdAt":      store.FormatTime(comment.CreatedAt),
		"updatedAt":      store.FormatTime(comment.UpdatedAt),
		"edited":         comment.Edited,
	}
}

func commentsJSON(comments []store.ServiceComment) []map[string]any {
	items := make([]map[string]any, 0, len(comments))
	for _, comment := range comments {
		items = append(items, commentJSON(comment))
	}
	return items
}
