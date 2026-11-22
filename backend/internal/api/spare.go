package api

import (
	"database/sql"
	"net/http"
	"strconv"

	"techservice/internal/store"
)

func listSpareParts(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query, ok := readSpareListQuery(w, r)
		if !ok {
			return
		}
		page, err := store.ListSpareParts(r.Context(), database, query)
		if err != nil {
			writeStoreError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"items": sparePartsJSON(page.Items), "hasMore": page.HasMore})
	}
}

func createSparePart(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Brand    string   `json:"brand"`
			Model    string   `json:"model"`
			Quantity *float64 `json:"quantity"`
		}
		if err := readJSON(w, r, &body); err != nil {
			writeError(w, http.StatusBadRequest, "invalid_json")
			return
		}
		created, err := store.CreateSparePart(r.Context(), database, userFrom(r.Context()), body.Brand, body.Model, body.Quantity)
		if err != nil {
			writeStoreError(w, err)
			return
		}
		payload := spareJSON(created.SparePart)
		payload["existed"] = created.Existed
		payload["added"] = created.Added
		writeJSON(w, http.StatusOK, payload)
	}
}

func getSparePart(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		part, err := store.GetSparePart(r.Context(), database, r.PathValue("id"))
		if err != nil {
			writeStoreError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, spareJSON(part))
	}
}

func updateSparePart(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Brand string `json:"brand"`
			Model string `json:"model"`
		}
		if err := readJSON(w, r, &body); err != nil {
			writeError(w, http.StatusBadRequest, "invalid_json")
			return
		}
		part, err := store.UpdateSparePart(r.Context(), database, userFrom(r.Context()), r.PathValue("id"), body.Brand, body.Model)
		if err != nil {
			writeStoreError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, spareJSON(part))
	}
}

func addSpareStock(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, ok := readStockBody(w, r)
		if !ok {
			return
		}
		part, err := store.AddSpareStock(r.Context(), database, userFrom(r.Context()), r.PathValue("id"), body.Quantity, body.Note)
		if err != nil {
			writeStoreError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, spareJSON(part))
	}
}

func writeOffSpareStock(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, ok := readStockBody(w, r)
		if !ok {
			return
		}
		part, err := store.WriteOffSpareStock(r.Context(), database, userFrom(r.Context()), r.PathValue("id"), body.Quantity, body.Note)
		if err != nil {
			writeStoreError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, spareJSON(part))
	}
}

func listSpareUses(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query, ok := readSpareMoveQuery(w, r)
		if !ok {
			return
		}
		page, err := store.ListSpareUses(r.Context(), database, r.PathValue("id"), query)
		if err != nil {
			writeStoreError(w, err)
			return
		}
		items := make([]map[string]any, 0, len(page.Items))
		for _, item := range page.Items {
			items = append(items, map[string]any{
				"id":            item.ID,
				"at":            store.FormatTime(item.At),
				"actorUsername": item.ActorUsername,
				"quantity":      item.Quantity,
				"kind":          item.Kind,
				"note":          item.Note,
				"status":        item.Status,
				"serviceId":     item.ServiceID,
				"deviceBrand":   item.DeviceBrand,
				"deviceModel":   item.DeviceModel,
				"serial":        item.Serial,
			})
		}
		writeJSON(w, http.StatusOK, map[string]any{"items": items, "hasMore": page.HasMore})
	}
}

func listSpareStock(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query, ok := readSpareMoveQuery(w, r)
		if !ok {
			return
		}
		page, err := store.ListSpareStock(r.Context(), database, r.PathValue("id"), query)
		if err != nil {
			writeStoreError(w, err)
			return
		}
		items := make([]map[string]any, 0, len(page.Items))
		for _, item := range page.Items {
			items = append(items, map[string]any{
				"id":            item.ID,
				"at":            store.FormatTime(item.At),
				"actorUsername": item.ActorUsername,
				"quantity":      item.Quantity,
				"kind":          item.Kind,
				"note":          item.Note,
			})
		}
		writeJSON(w, http.StatusOK, map[string]any{"items": items, "hasMore": page.HasMore})
	}
}

func listServiceSpares(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query, ok := readSpareMoveQuery(w, r)
		if !ok {
			return
		}
		page, err := store.ListServiceSpares(r.Context(), database, r.PathValue("id"), query)
		if err != nil {
			writeStoreError(w, err)
			return
		}
		items := make([]map[string]any, 0, len(page.Items))
		for _, item := range page.Items {
			items = append(items, map[string]any{
				"id":            item.ID,
				"at":            store.FormatTime(item.At),
				"actorUsername": item.ActorUsername,
				"quantity":      item.Quantity,
				"kind":          item.Kind,
				"note":          item.Note,
				"status":        item.Status,
				"partId":        item.PartID,
				"brand":         item.Brand,
				"model":         item.Model,
			})
		}
		writeJSON(w, http.StatusOK, map[string]any{"items": items, "hasMore": page.HasMore})
	}
}

func useSparePart(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Lines []struct {
				PartID   string   `json:"partId"`
				Quantity *float64 `json:"quantity"`
			} `json:"lines"`
		}
		if err := readJSON(w, r, &body); err != nil {
			writeError(w, http.StatusBadRequest, "invalid_json")
			return
		}
		lines := make([]store.SpareUseLine, 0, len(body.Lines))
		for _, line := range body.Lines {
			lines = append(lines, store.SpareUseLine{PartID: line.PartID, Quantity: line.Quantity})
		}
		if err := store.UseSpareParts(r.Context(), database, userFrom(r.Context()), r.PathValue("id"), lines); err != nil {
			writeStoreError(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func listReturnableSpares(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query, ok := readSpareListQuery(w, r)
		if !ok {
			return
		}
		page, err := store.ListReturnableSpares(r.Context(), database, r.PathValue("id"), query)
		if err != nil {
			writeStoreError(w, err)
			return
		}
		items := make([]map[string]any, 0, len(page.Items))
		for _, item := range page.Items {
			items = append(items, map[string]any{
				"id":         item.ID,
				"brand":      item.Brand,
				"model":      item.Model,
				"returnable": item.Returnable,
			})
		}
		writeJSON(w, http.StatusOK, map[string]any{"items": items, "hasMore": page.HasMore})
	}
}

func returnSparePart(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			PartID   string   `json:"partId"`
			Quantity *float64 `json:"quantity"`
			Note     string   `json:"note"`
		}
		if err := readJSON(w, r, &body); err != nil {
			writeError(w, http.StatusBadRequest, "invalid_json")
			return
		}
		onHand, err := store.ReturnSparePart(r.Context(), database, userFrom(r.Context()), r.PathValue("id"), body.PartID, body.Quantity, body.Note)
		if err != nil {
			writeStoreError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"onHand": onHand})
	}
}

func suggestSpareBrands(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		names, err := store.SuggestSpareBrands(r.Context(), database, r.URL.Query().Get("q"))
		if err != nil {
			writeStoreError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"items": names})
	}
}

func suggestSpareModels(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		names, err := store.SuggestSpareModels(r.Context(), database, r.URL.Query().Get("brand"), r.URL.Query().Get("q"))
		if err != nil {
			writeStoreError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"items": names})
	}
}

func spareJSON(part store.SparePart) map[string]any {
	return map[string]any{
		"id":     part.ID,
		"brand":  part.Brand,
		"model":  part.Model,
		"onHand": part.OnHand,
	}
}

func sparePartsJSON(parts []store.SparePart) []map[string]any {
	items := make([]map[string]any, 0, len(parts))
	for _, part := range parts {
		items = append(items, spareJSON(part))
	}
	return items
}

type stockBody struct {
	Quantity *float64 `json:"quantity"`
	Note     string   `json:"note"`
}

func readStockBody(w http.ResponseWriter, r *http.Request) (stockBody, bool) {
	var body stockBody
	if err := readJSON(w, r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json")
		return stockBody{}, false
	}
	return body, true
}

func readSpareListQuery(w http.ResponseWriter, r *http.Request) (store.SpareListQuery, bool) {
	limit, err := queryLimit(r)
	if err != nil {
		writeStoreError(w, err)
		return store.SpareListQuery{}, false
	}
	query := store.SpareListQuery{Text: r.URL.Query().Get("q"), Limit: limit}
	brand := r.URL.Query().Get("after")
	model := r.URL.Query().Get("afterModel")
	id := r.URL.Query().Get("afterId")
	if (brand == "") != (model == "") || (brand == "") != (id == "") {
		writeError(w, http.StatusBadRequest, "invalid_cursor")
		return store.SpareListQuery{}, false
	}
	if brand != "" {
		query.HasAfter = true
		query.AfterBrand = brand
		query.AfterModel = model
		query.AfterID = id
	}
	return query, true
}

func readSpareMoveQuery(w http.ResponseWriter, r *http.Request) (store.SpareMoveQuery, bool) {
	limit, err := queryLimit(r)
	if err != nil {
		writeStoreError(w, err)
		return store.SpareMoveQuery{}, false
	}
	query := store.SpareMoveQuery{Limit: limit}
	after := r.URL.Query().Get("after")
	afterID := r.URL.Query().Get("afterId")
	if (after == "") != (afterID == "") {
		writeError(w, http.StatusBadRequest, "invalid_cursor")
		return store.SpareMoveQuery{}, false
	}
	if after != "" {
		id, err := strconv.ParseInt(afterID, 10, 64)
		if err != nil || id < 1 {
			writeError(w, http.StatusBadRequest, "invalid_cursor")
			return store.SpareMoveQuery{}, false
		}
		query.HasAfter = true
		query.AfterTime = after
		query.AfterID = id
	}
	return query, true
}
