package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"techservice/internal/store"
)

func listCustomers(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		limit, err := queryLimit(r)
		if err != nil {
			writeStoreError(w, err)
			return
		}
		query := store.CustomerQuery{Text: r.URL.Query().Get("q"), Limit: limit}
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
			query.AfterName = after
			query.AfterID = id
		}
		page, err := store.ListCustomers(r.Context(), database, query)
		if err != nil {
			writeStoreError(w, err)
			return
		}
		items := make([]map[string]any, 0, len(page.Items))
		for _, customer := range page.Items {
			items = append(items, customerJSON(customer))
		}
		writeJSON(w, http.StatusOK, map[string]any{"items": items, "hasMore": page.HasMore})
	}
}

func getCustomer(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid_id")
			return
		}
		customer, err := store.GetCustomer(r.Context(), database, id)
		if err != nil {
			writeStoreError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, customerJSON(customer))
	}
}

func updateCustomer(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid_id")
			return
		}
		var body struct {
			Name  string `json:"name"`
			Phone string `json:"phone"`
			Email string `json:"email"`
		}
		if err := readJSON(w, r, &body); err != nil {
			writeError(w, http.StatusBadRequest, "invalid_json")
			return
		}
		customer, err := store.UpdateCustomer(r.Context(), database, userFrom(r.Context()), id, body.Name, body.Phone, body.Email)
		if err != nil {
			writeStoreError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, customerJSON(customer))
	}
}

func listDevices(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		limit, err := queryLimit(r)
		if err != nil {
			writeStoreError(w, err)
			return
		}
		query := store.DeviceQuery{Text: r.URL.Query().Get("q"), Status: r.URL.Query().Get("status"), Limit: limit}
		if value := r.URL.Query().Get("customerId"); value != "" {
			id, err := strconv.ParseInt(value, 10, 64)
			if err != nil || id <= 0 {
				writeError(w, http.StatusBadRequest, "invalid_id")
				return
			}
			query.HasCustomer = true
			query.CustomerID = id
		}
		after := r.URL.Query().Get("after")
		afterID := r.URL.Query().Get("afterId")
		if (after == "") != (afterID == "") {
			writeError(w, http.StatusBadRequest, "invalid_cursor")
			return
		}
		if after != "" {
			query.HasAfter = true
			query.AfterTime = after
			query.AfterID = afterID
		}
		page, err := store.ListDevices(r.Context(), database, query)
		if err != nil {
			writeStoreError(w, err)
			return
		}
		items := make([]map[string]any, 0, len(page.Items))
		for _, item := range page.Items {
			items = append(items, map[string]any{
				"id":           item.ID,
				"name":         item.Name,
				"brand":        item.Brand,
				"model":        item.Model,
				"customerName": item.CustomerName,
				"status":       item.Status,
				"createdAt":    store.FormatTime(item.CreatedAt),
			})
		}
		writeJSON(w, http.StatusOK, map[string]any{"items": items, "hasMore": page.HasMore})
	}
}

func getDevice(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		device, err := store.GetDevice(r.Context(), database, r.PathValue("id"))
		if err != nil {
			writeStoreError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, deviceJSON(device))
	}
}

func registerDevice(database *sql.DB, root string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			CustomerID    *int64 `json:"customerId"`
			CustomerName  string `json:"customerName"`
			CustomerPhone string `json:"customerPhone"`
			CustomerEmail string `json:"customerEmail"`
			Name          string `json:"name"`
			Brand         string `json:"brand"`
			Model         string `json:"model"`
			Serial        string `json:"serial"`
			Note          string `json:"note"`
			Parts         []struct {
				Name   string `json:"name"`
				Note   string `json:"note"`
				Photos []int  `json:"photos"`
			} `json:"parts"`
			DevicePhotos []int `json:"devicePhotos"`
		}
		files, err := readDeviceForm(w, r, &body)
		if err != nil {
			writeFormError(w, err)
			return
		}
		defer clearForm(r)
		parts := make([]store.NewPart, len(body.Parts))
		for i, part := range body.Parts {
			parts[i] = store.NewPart{Name: part.Name, Note: part.Note, Photos: part.Photos}
		}
		device, err := store.RegisterDevice(r.Context(), database, root, userFrom(r.Context()), store.NewDevice{
			CustomerID:    body.CustomerID,
			CustomerName:  body.CustomerName,
			CustomerPhone: body.CustomerPhone,
			CustomerEmail: body.CustomerEmail,
			Name:          body.Name,
			Brand:         body.Brand,
			Model:         body.Model,
			Serial:        body.Serial,
			Note:          body.Note,
			Parts:         parts,
			Files:         files,
			DevicePhotos:  body.DevicePhotos,
		})
		if err != nil {
			writeStoreError(w, err)
			return
		}
		writeJSON(w, http.StatusCreated, deviceJSON(device))
	}
}

func updateDevice(database *sql.DB, root string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Name          string `json:"name"`
			Brand         string `json:"brand"`
			Model         string `json:"model"`
			Serial        string `json:"serial"`
			Note          string `json:"note"`
			CustomerName  string `json:"customerName"`
			CustomerPhone string `json:"customerPhone"`
			CustomerEmail string `json:"customerEmail"`
			Parts         []struct {
				ID     string   `json:"id"`
				Name   string   `json:"name"`
				Note   string   `json:"note"`
				Keep   []string `json:"keep"`
				Photos []int    `json:"photos"`
			} `json:"parts"`
			DeviceKeep   []string `json:"deviceKeep"`
			DevicePhotos []int    `json:"devicePhotos"`
		}
		files, err := readDeviceForm(w, r, &body)
		if err != nil {
			writeFormError(w, err)
			return
		}
		defer clearForm(r)
		parts := make([]store.PartChange, len(body.Parts))
		for i, part := range body.Parts {
			parts[i] = store.PartChange{ID: part.ID, Name: part.Name, Note: part.Note, Keep: part.Keep, Photos: part.Photos}
		}
		device, err := store.UpdateDevice(r.Context(), database, root, userFrom(r.Context()), r.PathValue("id"), store.DeviceChange{
			Name:          body.Name,
			Brand:         body.Brand,
			Model:         body.Model,
			Serial:        body.Serial,
			Note:          body.Note,
			CustomerName:  body.CustomerName,
			CustomerPhone: body.CustomerPhone,
			CustomerEmail: body.CustomerEmail,
			Parts:         parts,
			Files:         files,
			DeviceKeep:    body.DeviceKeep,
			DevicePhotos:  body.DevicePhotos,
		})
		if err != nil {
			writeStoreError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, deviceJSON(device))
	}
}

func changeDeviceStatus(database *sql.DB, root string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Status string `json:"status"`
			Note   string `json:"note"`
		}
		file, err := readStatusForm(w, r, &body)
		if err != nil {
			writeFormError(w, err)
			return
		}
		defer clearForm(r)
		device, err := store.ChangeDeviceStatus(r.Context(), database, root, userFrom(r.Context()), r.PathValue("id"), body.Status, body.Note, file)
		if err != nil {
			writeStoreError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, deviceJSON(device))
	}
}

func getStatusPhoto(database *sql.DB, root string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		changeID, err := strconv.ParseInt(r.PathValue("changeId"), 10, 64)
		if err != nil || changeID <= 0 {
			writeError(w, http.StatusBadRequest, "invalid_id")
			return
		}
		full, media, err := store.OpenStatusPhoto(r.Context(), database, root, r.PathValue("id"), changeID)
		if err != nil {
			writeStoreError(w, err)
			return
		}
		file, err := os.Open(full)
		if err != nil {
			writeStoreError(w, err)
			return
		}
		defer file.Close()
		w.Header().Set("Content-Type", media)
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Cache-Control", "private, max-age=3600")
		http.ServeContent(w, r, "", time.Time{}, file)
	}
}

func getPhoto(database *sql.DB, root string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		full, media, err := store.OpenPhoto(r.Context(), database, root, r.PathValue("id"), r.PathValue("photoId"))
		if err != nil {
			writeStoreError(w, err)
			return
		}
		file, err := os.Open(full)
		if err != nil {
			writeStoreError(w, err)
			return
		}
		defer file.Close()
		w.Header().Set("Content-Type", media)
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Cache-Control", "private, max-age=3600")
		http.ServeContent(w, r, "", time.Time{}, file)
	}
}

func customerJSON(customer store.Customer) map[string]any {
	return map[string]any{
		"id":    customer.ID,
		"name":  customer.Name,
		"phone": customer.Phone,
		"email": customer.Email,
	}
}

func deviceJSON(device store.Device) map[string]any {
	parts := make([]map[string]any, 0, len(device.Parts))
	for _, part := range device.Parts {
		parts = append(parts, map[string]any{
			"id":     part.ID,
			"name":   part.Name,
			"note":   part.Note,
			"photos": photosJSON(part.Photos),
		})
	}
	return map[string]any{
		"id":        device.ID,
		"name":      device.Name,
		"brand":     device.Brand,
		"model":     device.Model,
		"serial":    device.Serial,
		"note":      device.Note,
		"status":    device.Status,
		"createdAt": store.FormatTime(device.CreatedAt),
		"customer":  customerJSON(device.Customer),
		"parts":     parts,
		"photos":    photosJSON(device.Photos),
		"history":   historyJSON(device.History),
	}
}

func historyJSON(history []store.StatusChange) []map[string]any {
	items := make([]map[string]any, 0, len(history))
	for _, change := range history {
		items = append(items, map[string]any{
			"id":            change.ID,
			"actorUsername": change.ActorUsername,
			"from":          change.From,
			"to":            change.To,
			"note":          change.Note,
			"hasPhoto":      change.HasPhoto,
			"createdAt":     store.FormatTime(change.CreatedAt),
		})
	}
	return items
}

func photosJSON(photos []store.Photo) []map[string]any {
	items := make([]map[string]any, 0, len(photos))
	for _, photo := range photos {
		items = append(items, map[string]any{
			"id":        photo.ID,
			"mediaType": photo.MediaType,
			"size":      photo.Size,
		})
	}
	return items
}

func readStatusForm(w http.ResponseWriter, r *http.Request, dest any) (*store.Upload, error) {
	r.Body = http.MaxBytesReader(w, r.Body, store.StatusUploadBody)
	if err := r.ParseMultipartForm(1 << 20); err != nil {
		var max *http.MaxBytesError
		if errors.As(err, &max) {
			return nil, &store.FieldError{Photo: "size"}
		}
		return nil, err
	}
	decoder := json.NewDecoder(strings.NewReader(r.FormValue("payload")))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(dest); err != nil {
		return nil, err
	}
	for name := range r.MultipartForm.File {
		if name != "file0" {
			return nil, &store.FieldError{Invalid: []string{"statusPhoto"}}
		}
	}
	header, err := formFile(r, "file0")
	if errors.Is(err, http.ErrMissingFile) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &store.Upload{Open: func() (io.ReadCloser, error) {
		return header.Open()
	}}, nil
}

func readDeviceForm(w http.ResponseWriter, r *http.Request, dest any) ([]store.Upload, error) {
	r.Body = http.MaxBytesReader(w, r.Body, store.MaxUploadBody)
	if err := r.ParseMultipartForm(8 << 20); err != nil {
		var max *http.MaxBytesError
		if errors.As(err, &max) {
			return nil, &store.FieldError{Photo: "size"}
		}
		return nil, err
	}
	decoder := json.NewDecoder(strings.NewReader(r.FormValue("payload")))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(dest); err != nil {
		return nil, err
	}
	var uploads []store.Upload
	for i := 0; ; i++ {
		header, err := formFile(r, "file"+strconv.Itoa(i))
		if errors.Is(err, http.ErrMissingFile) {
			break
		}
		if err != nil {
			return nil, err
		}
		uploads = append(uploads, store.Upload{Open: func() (io.ReadCloser, error) {
			return header.Open()
		}})
	}
	return uploads, nil
}

func formFile(r *http.Request, name string) (*multipart.FileHeader, error) {
	if r.MultipartForm == nil || r.MultipartForm.File == nil {
		return nil, http.ErrMissingFile
	}
	headers := r.MultipartForm.File[name]
	if len(headers) == 0 {
		return nil, http.ErrMissingFile
	}
	return headers[0], nil
}

func clearForm(r *http.Request) {
	if r.MultipartForm != nil {
		r.MultipartForm.RemoveAll()
	}
}

func writeFormError(w http.ResponseWriter, err error) {
	var fields *store.FieldError
	if errors.As(err, &fields) {
		writeValidation(w, fields)
		return
	}
	writeError(w, http.StatusBadRequest, "invalid_form")
}

func queryLimit(r *http.Request) (int, error) {
	value := r.URL.Query().Get("limit")
	if value == "" {
		return 0, nil
	}
	limit, err := strconv.Atoi(value)
	if err != nil {
		return 0, store.ErrInvalidLimit
	}
	return limit, nil
}
