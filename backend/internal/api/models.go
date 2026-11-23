package api

import (
	"context"
	"database/sql"
	"net/http"

	"techservice/internal/catalog"
	"techservice/internal/store"
)

func listModels(database *sql.DB, files catalog.Files) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query, ok := readModelListQuery(w, r)
		if !ok {
			return
		}
		page, err := store.ListCatalogModels(r.Context(), database, query)
		if err != nil {
			writeStoreError(w, err)
			return
		}
		items := make([]map[string]any, 0, len(page.Items))
		for _, item := range page.Items {
			items = append(items, modelJSON(item, files.Has(item.Brand, item.Model)))
		}
		writeJSON(w, http.StatusOK, map[string]any{"items": items, "hasMore": page.HasMore})
	}
}

func createModel(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Brand string `json:"brand"`
			Model string `json:"model"`
		}
		if err := readJSON(w, r, &body); err != nil {
			writeError(w, http.StatusBadRequest, "invalid_json")
			return
		}
		item, err := store.CreateCatalogModel(r.Context(), database, userFrom(r.Context()), body.Brand, body.Model)
		if err != nil {
			writeStoreError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, modelJSON(item, false))
	}
}

func updateModel(database *sql.DB, files catalog.Files) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Active *bool `json:"active"`
		}
		if err := readJSON(w, r, &body); err != nil {
			writeError(w, http.StatusBadRequest, "invalid_json")
			return
		}
		if body.Active == nil {
			writeStoreError(w, &store.FieldError{Invalid: []string{"active"}})
			return
		}
		item, err := store.SetCatalogActive(r.Context(), database, userFrom(r.Context()), r.PathValue("id"), *body.Active)
		if err != nil {
			writeStoreError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, modelJSON(item, files.Has(item.Brand, item.Model)))
	}
}

func suggestModelBrands(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		activeOnly, ok := modelActiveQuery(w, r)
		if !ok {
			return
		}
		names, err := store.SuggestCatalogBrands(r.Context(), database, r.URL.Query().Get("q"), activeOnly)
		if err != nil {
			writeStoreError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"items": names})
	}
}

func suggestModelNames(database *sql.DB, files catalog.Files) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		activeOnly, ok := modelActiveQuery(w, r)
		if !ok {
			return
		}
		pairs, err := store.SuggestCatalogModels(r.Context(), database, r.URL.Query().Get("brand"), r.URL.Query().Get("q"), activeOnly)
		if err != nil {
			writeStoreError(w, err)
			return
		}
		items := make([]map[string]any, 0, len(pairs))
		for _, pair := range pairs {
			items = append(items, map[string]any{
				"name":  pair.Model,
				"image": files.Has(pair.Brand, pair.Model),
			})
		}
		writeJSON(w, http.StatusOK, map[string]any{"items": items})
	}
}

func modelActiveQuery(w http.ResponseWriter, r *http.Request) (bool, bool) {
	if !r.URL.Query().Has("active") {
		return false, true
	}
	if r.URL.Query().Get("active") != "1" {
		writeError(w, http.StatusBadRequest, "invalid_active")
		return false, false
	}
	return true, true
}

func readModelListQuery(w http.ResponseWriter, r *http.Request) (store.CatalogListQuery, bool) {
	limit, err := queryLimit(r)
	if err != nil {
		writeStoreError(w, err)
		return store.CatalogListQuery{}, false
	}
	query := store.CatalogListQuery{Text: r.URL.Query().Get("q"), Limit: limit}
	brand := r.URL.Query().Get("after")
	model := r.URL.Query().Get("afterModel")
	id := r.URL.Query().Get("afterId")
	if (brand == "") != (model == "") || (brand == "") != (id == "") {
		writeError(w, http.StatusBadRequest, "invalid_cursor")
		return store.CatalogListQuery{}, false
	}
	if brand != "" {
		query.HasAfter = true
		query.AfterBrand = brand
		query.AfterModel = model
		query.AfterID = id
	}
	return query, true
}

func modelJSON(item store.CatalogModel, image bool) map[string]any {
	return map[string]any{
		"id":     item.ID,
		"brand":  item.Brand,
		"model":  item.Model,
		"source": item.Source,
		"active": item.Active,
		"image":  image,
	}
}

func catalogImage(ctx context.Context, database *sql.DB, files catalog.Files, brand, model string) (bool, error) {
	flags, err := catalogImages(ctx, database, files, []store.CatalogPair{{Brand: brand, Model: model}})
	if err != nil {
		return false, err
	}
	return flags[store.CatalogKey(brand, model)], nil
}

func catalogImages(ctx context.Context, database *sql.DB, files catalog.Files, pairs []store.CatalogPair) (map[string]bool, error) {
	canon, err := store.CatalogCanon(ctx, database, pairs)
	if err != nil {
		return nil, err
	}
	flags := map[string]bool{}
	for _, pair := range pairs {
		key := store.CatalogKey(pair.Brand, pair.Model)
		name, ok := canon[key]
		if !ok {
			flags[key] = false
			continue
		}
		flags[key] = files.Has(name.Brand, name.Model)
	}
	return flags, nil
}
