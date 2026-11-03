package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"techservice/internal/store"
)

func listAudit(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		limit, err := queryLimit(r)
		if err != nil {
			writeStoreError(w, err)
			return
		}
		from, err := queryTime(r.URL.Query().Get("from"))
		if err != nil {
			writeStoreError(w, err)
			return
		}
		to, err := queryTime(r.URL.Query().Get("to"))
		if err != nil {
			writeStoreError(w, err)
			return
		}
		query := store.AuditQuery{
			ViewerRole: userFrom(r.Context()).Role,
			From:       from,
			To:         to,
			Person:     r.URL.Query().Get("person"),
			Event:      r.URL.Query().Get("event"),
			Subject:    r.URL.Query().Get("subject"),
			Sort:       r.URL.Query().Get("sort"),
			Limit:      limit,
		}
		after := r.URL.Query().Get("after")
		afterID := r.URL.Query().Get("afterId")
		if (after == "") != (afterID == "") {
			writeError(w, http.StatusBadRequest, "invalid_cursor")
			return
		}
		if after != "" {
			id, err := strconv.ParseInt(afterID, 10, 64)
			if err != nil {
				writeError(w, http.StatusBadRequest, "invalid_cursor")
				return
			}
			query.HasAfter = true
			query.AfterTime = after
			query.AfterID = id
		}
		page, err := store.ListAudit(r.Context(), database, query)
		if err != nil {
			writeStoreError(w, err)
			return
		}
		items := make([]map[string]any, 0, len(page.Items))
		for _, item := range page.Items {
			detail := item.Detail
			if len(detail) == 0 {
				detail = json.RawMessage(`{}`)
			}
			items = append(items, map[string]any{
				"id":            item.ID,
				"at":            store.FormatTime(item.At),
				"actorUsername": item.ActorUsername,
				"actorRole":     item.ActorRole,
				"event":         item.Event,
				"detail":        detail,
			})
		}
		writeJSON(w, http.StatusOK, map[string]any{"items": items, "hasMore": page.HasMore})
	}
}

func queryTime(value string) (*time.Time, error) {
	if value == "" {
		return nil, nil
	}
	parsed, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		return nil, store.ErrInvalidTime
	}
	utc := parsed.UTC()
	return &utc, nil
}
