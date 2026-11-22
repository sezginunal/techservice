package api

import (
	"archive/zip"
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"techservice/internal/brand"
	"techservice/internal/catalog"
	"techservice/internal/db"
	"techservice/internal/store"
)

func TestExportCustomersRequiresManager(t *testing.T) {
	database := openAPITestDB(t)
	handler := New(database, t.TempDir(), brand.Brand{Name: "Shop"}, catalog.Index{}, "en")

	techCookie := loginCookie(t, handler, "tech", "tech12345")
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/customers/export?lang=en", nil)
	req.AddCookie(techCookie)
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("technician status %d body %s", rec.Code, rec.Body.String())
	}

	ownerCookie := loginCookie(t, handler, "owner", "owner1234")
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/customers/export?lang=en&q=Ada", nil)
	req.AddCookie(ownerCookie)
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("owner status %d body %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Header().Get("Content-Type"), "spreadsheetml") {
		t.Fatalf("content type %q", rec.Header().Get("Content-Type"))
	}
	shared := zipSharedStrings(t, rec.Body.Bytes())
	if !strings.Contains(shared, "Ada") || !strings.Contains(shared, "Name") {
		t.Fatalf("shared strings %s", shared)
	}
}

func TestExportServicesUsesFilterAndLanguage(t *testing.T) {
	database := openAPITestDB(t)
	handler := New(database, t.TempDir(), brand.Brand{Name: "Shop"}, catalog.Index{}, "en")
	cookie := loginCookie(t, handler, "admin", "admin1234")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/services/export?lang=tr&status=registered", nil)
	req.AddCookie(cookie)
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d body %s", rec.Code, rec.Body.String())
	}
	shared := zipSharedStrings(t, rec.Body.Bytes())
	if !strings.Contains(shared, "Servisler") && !strings.Contains(shared, "Kayıt yapıldı") {
		// sheet name is in workbook; status label is in shared strings
	}
	if !strings.Contains(shared, "Kayıt yapıldı") {
		t.Fatalf("expected Turkish status label in %s", shared)
	}
	if strings.Contains(shared, "Delivered") {
		t.Fatalf("filter leaked delivered row: %s", shared)
	}
}

func openAPITestDB(t *testing.T) *sql.DB {
	t.Helper()
	ctx := context.Background()
	database, err := db.Open(ctx, t.TempDir()+"/test.db")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { database.Close() })
	if err := store.Migrate(ctx, database); err != nil {
		t.Fatal(err)
	}
	if err := store.EnsureAdmin(ctx, database, "admin", "admin1234"); err != nil {
		t.Fatal(err)
	}
	admin, err := store.Login(ctx, database, "admin", "admin1234")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.CreateUser(ctx, database, admin.User, "owner", "owner1234", store.RoleOwner); err != nil {
		t.Fatal(err)
	}
	if _, err := store.CreateUser(ctx, database, admin.User, "tech", "tech12345", store.RoleTechnician); err != nil {
		t.Fatal(err)
	}
	root := t.TempDir()
	if _, err := store.RegisterDevice(ctx, database, root, admin.User, store.NewDevice{
		Brand:         "Epson",
		Model:         "L3250",
		Serial:        "SN-1",
		CustomerName:  "Ada",
		CustomerPhone: "5551112233",
	}); err != nil {
		t.Fatal(err)
	}
	delivered, err := store.RegisterDevice(ctx, database, root, admin.User, store.NewDevice{
		Brand:         "Canon",
		Model:         "G3010",
		Serial:        "SN-2",
		CustomerName:  "Bob",
		CustomerPhone: "5554445566",
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.ChangeDeviceStatus(ctx, database, root, admin.User, delivered.ID, store.StatusDelivered, "", nil); err != nil {
		t.Fatal(err)
	}
	return database
}

func loginCookie(t *testing.T, handler http.Handler, username, password string) *http.Cookie {
	t.Helper()
	body, _ := json.Marshal(map[string]string{"username": username, "password": password})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("login %s: %d %s", username, rec.Code, rec.Body.String())
	}
	for _, cookie := range rec.Result().Cookies() {
		if cookie.Name == sessionCookie {
			return cookie
		}
	}
	t.Fatal("missing session cookie")
	return nil
}

func zipSharedStrings(t *testing.T, raw []byte) string {
	t.Helper()
	reader, err := zip.NewReader(bytes.NewReader(raw), int64(len(raw)))
	if err != nil {
		t.Fatal(err)
	}
	for _, file := range reader.File {
		if file.Name != "xl/sharedStrings.xml" {
			continue
		}
		part, err := file.Open()
		if err != nil {
			t.Fatal(err)
		}
		defer part.Close()
		body, err := io.ReadAll(part)
		if err != nil {
			t.Fatal(err)
		}
		return string(body)
	}
	t.Fatal("missing shared strings")
	return ""
}
