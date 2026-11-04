package store

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"techservice/internal/db"
)

func TestIntakeAndAudit(t *testing.T) {
	ctx := context.Background()
	database := openTestDB(t)
	root := t.TempDir()

	if err := EnsureAdmin(ctx, database, "admin", "admin1234"); err != nil {
		t.Fatal(err)
	}
	admin, err := Login(ctx, database, "admin", "admin1234")
	if err != nil {
		t.Fatal(err)
	}
	owner, err := CreateUser(ctx, database, admin.User, "owner", "owner1234", RoleOwner)
	if err != nil {
		t.Fatal(err)
	}
	tech, err := CreateUser(ctx, database, admin.User, "tech", "tech12345", RoleTechnician)
	if err != nil {
		t.Fatal(err)
	}

	_, err = RegisterDevice(ctx, database, root, tech, NewDevice{CustomerName: "Ada"})
	var fields *FieldError
	if !errors.As(err, &fields) || !contains(fields.Missing, "brand") || contains(fields.Missing, "name") {
		t.Fatalf("missing brand: %v", err)
	}
	if countRows(t, database, `SELECT COUNT(*) FROM customers`) != 0 {
		t.Fatal("invalid register wrote a customer")
	}

	_, err = RegisterDevice(ctx, database, root, tech, NewDevice{
		CustomerName: "Ada", Brand: "HP", Model: "Laser", Note: "ok", CustomerEmail: "nope",
	})
	if !errors.As(err, &fields) || !contains(fields.Invalid, "email") {
		t.Fatalf("email: %v", err)
	}
	_, err = RegisterDevice(ctx, database, root, tech, NewDevice{
		CustomerName: "Ada", Brand: "HP", Model: "Laser", Note: "ok", CustomerPhone: "asdas",
	})
	if !errors.As(err, &fields) || !contains(fields.Invalid, "phone") {
		t.Fatalf("phone: %v", err)
	}
	_, err = RegisterDevice(ctx, database, root, tech, NewDevice{
		CustomerName: "Ada", Brand: "HP", Model: "Laser", Note: "ok", CustomerPhone: "555",
	})
	if !errors.As(err, &fields) || !contains(fields.Invalid, "phone") {
		t.Fatalf("short phone: %v", err)
	}

	device, err := RegisterDevice(ctx, database, root, tech, NewDevice{
		CustomerName:  "Ada Lovelace",
		CustomerPhone: "555 111 22 33",
		Brand:         "HP",
		Model:         "Laser",
		Note:          "works",
		Parts:         []NewPart{{Name: "Cable", Note: "ok", Photos: []int{0}}},
		Files:         []Upload{uploadBytes(pngBytes())},
	})
	if err != nil {
		t.Fatal(err)
	}
	if device.Status != StatusRegistered || device.Customer.Name != "Ada Lovelace" || len(device.Parts) != 1 || len(device.Parts[0].Photos) != 1 {
		t.Fatalf("device %+v", device)
	}
	if _, err := os.Stat(filepath.Join(root, device.Parts[0].Photos[0].Path)); err != nil {
		t.Fatal(err)
	}
	if countRows(t, database, `SELECT COUNT(*) FROM audit_events WHERE event = ?`, EventCustomerCreated) != 1 {
		t.Fatal("customer audit")
	}
	if countRows(t, database, `SELECT COUNT(*) FROM audit_events WHERE event = ?`, EventDeviceRegistered) != 1 {
		t.Fatal("device audit")
	}

	found, err := ListDevices(ctx, database, DeviceQuery{Text: "555", Limit: 10})
	if err != nil || len(found.Items) != 1 || found.Items[0].ID != device.ID {
		t.Fatalf("search %+v %v", found, err)
	}
	byID, err := ListDevices(ctx, database, DeviceQuery{Text: device.ID, Limit: 10})
	if err != nil || len(byID.Items) != 1 {
		t.Fatalf("uuid search %+v %v", byID, err)
	}

	ownerPage, err := ListAudit(ctx, database, AuditQuery{ViewerRole: owner.Role, Limit: 20})
	if err != nil {
		t.Fatal(err)
	}
	for _, item := range ownerPage.Items {
		if item.Event == EventUserCreated || item.ActorRole == RoleAdmin {
			t.Fatalf("owner saw %s by %s", item.Event, item.ActorRole)
		}
	}
	if _, err := ListAudit(ctx, database, AuditQuery{ViewerRole: RoleTechnician}); !errors.Is(err, ErrForbidden) {
		t.Fatalf("technician audit: %v", err)
	}
	adminPage, err := ListAudit(ctx, database, AuditQuery{ViewerRole: RoleAdmin, Event: EventUserCreated, Limit: 20})
	if err != nil || len(adminPage.Items) != 2 {
		t.Fatalf("admin users %+v %v", adminPage, err)
	}

	renamed, err := UpdateDevice(ctx, database, root, tech, device.ID, DeviceChange{
		Brand: device.Brand, Model: device.Model, Note: device.Note,
		CustomerName: "Ada King", CustomerPhone: "555 111 22 33",
		Parts: []PartChange{{
			ID: device.Parts[0].ID, Name: "Cable", Note: "ok", Keep: []string{device.Parts[0].Photos[0].ID},
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if renamed.Customer.Name != "Ada King" {
		t.Fatal(renamed.Customer.Name)
	}
	if countRows(t, database, `SELECT COUNT(*) FROM audit_events WHERE event = ?`, EventDeviceUpdated) != 0 {
		t.Fatal("customer-only save wrote a device update")
	}
	renames, err := ListAudit(ctx, database, AuditQuery{ViewerRole: RoleAdmin, Subject: "Lovelace", Event: EventCustomerUpdated, Limit: 10})
	if err != nil || len(renames.Items) != 1 || !bytes.Contains(renames.Items[0].Detail, []byte("Ada King")) {
		t.Fatalf("rename audit %+v %v", renames, err)
	}

	same, err := UpdateDevice(ctx, database, root, tech, device.ID, DeviceChange{
		Brand: renamed.Brand, Model: renamed.Model, Note: renamed.Note,
		CustomerName: renamed.Customer.Name, CustomerPhone: renamed.Customer.Phone,
		Parts: []PartChange{{
			ID: renamed.Parts[0].ID, Name: renamed.Parts[0].Name, Note: renamed.Parts[0].Note,
			Keep: []string{renamed.Parts[0].Photos[0].ID},
		}},
	})
	if err != nil || same.Note != renamed.Note {
		t.Fatal(err)
	}
	if countRows(t, database, `SELECT COUNT(*) FROM audit_events WHERE event = ?`, EventCustomerUpdated) != 1 {
		t.Fatal("unchanged save wrote audit")
	}

	_, err = UpdateDevice(ctx, database, root, tech, device.ID, DeviceChange{
		Brand: renamed.Brand, Model: renamed.Model, Note: "changed",
		CustomerName: renamed.Customer.Name, CustomerPhone: renamed.Customer.Phone,
		Parts: []PartChange{{
			ID: renamed.Parts[0].ID, Name: renamed.Parts[0].Name, Note: renamed.Parts[0].Note,
			Keep: []string{renamed.Parts[0].Photos[0].ID},
		}},
		Files:        []Upload{uploadBytes([]byte("nope"))},
		DevicePhotos: []int{0},
	})
	if !errors.As(err, &fields) || fields.Photo != "type" {
		t.Fatalf("bad photo: %v", err)
	}
	kept, err := GetDevice(ctx, database, device.ID)
	if err != nil || kept.Note != "works" {
		t.Fatalf("note changed on failed save: %+v %v", kept, err)
	}

	cleared, err := UpdateDevice(ctx, database, root, tech, device.ID, DeviceChange{
		Brand: kept.Brand, Model: kept.Model, Note: kept.Note,
		CustomerName: kept.Customer.Name, CustomerPhone: kept.Customer.Phone,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(cleared.Parts) != 0 {
		t.Fatal("part remained")
	}
	if _, err := os.Stat(filepath.Join(root, kept.Parts[0].Photos[0].Path)); !os.IsNotExist(err) {
		t.Fatalf("photo file: %v", err)
	}

	big := make([]byte, maxPhotoBytes+1)
	copy(big, pngBytes())
	_, err = RegisterDevice(ctx, database, root, tech, NewDevice{
		CustomerName: "Big", Brand: "Big", Model: "Big", Note: "big",
		Files: []Upload{uploadBytes(big)}, DevicePhotos: []int{0},
	})
	if !errors.As(err, &fields) || fields.Photo != "size" {
		t.Fatalf("size: %v", err)
	}

	files := make([]Upload, maxPhotos+1)
	indexes := make([]int, maxPhotos+1)
	for i := range files {
		files[i] = uploadBytes(pngBytes())
		indexes[i] = i
	}
	_, err = RegisterDevice(ctx, database, root, tech, NewDevice{
		CustomerName: "Many", Brand: "Many", Model: "Many", Note: "many",
		Files: files, DevicePhotos: indexes,
	})
	if !errors.As(err, &fields) || fields.Photo != "limit" {
		t.Fatalf("limit: %v", err)
	}

	for _, model := range []string{"First", "Second", "Third"} {
		if _, err := RegisterDevice(ctx, database, root, owner, NewDevice{
			CustomerID: &kept.Customer.ID, Brand: "HP", Model: model, Note: "n",
		}); err != nil {
			t.Fatal(err)
		}
	}
	first, err := ListDevices(ctx, database, DeviceQuery{Limit: 2})
	if err != nil || len(first.Items) != 2 || !first.HasMore {
		t.Fatalf("page %+v %v", first, err)
	}
	second, err := ListDevices(ctx, database, DeviceQuery{
		Limit: 2, HasAfter: true, AfterTime: timestamp(first.Items[1].CreatedAt), AfterID: first.Items[1].ID,
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, item := range second.Items {
		if item.ID == first.Items[0].ID || item.ID == first.Items[1].ID {
			t.Fatalf("page repeated %s", item.ID)
		}
	}

	password := "changed-password"
	if _, err := UpdateUser(ctx, database, admin.User, tech.ID, UserPatch{Password: &password}); err != nil {
		t.Fatal(err)
	}
	var detail string
	if err := database.QueryRowContext(ctx, `
		SELECT detail FROM audit_events WHERE event = ? ORDER BY id DESC LIMIT 1
	`, EventUserPasswordChanged).Scan(&detail); err != nil {
		t.Fatal(err)
	}
	if bytes.Contains([]byte(detail), []byte(password)) {
		t.Fatal("password stored in audit")
	}
}

func TestDeviceStatus(t *testing.T) {
	ctx := context.Background()
	database := openTestDB(t)
	root := t.TempDir()
	if err := EnsureAdmin(ctx, database, "admin", "admin1234"); err != nil {
		t.Fatal(err)
	}
	admin, err := Login(ctx, database, "admin", "admin1234")
	if err != nil {
		t.Fatal(err)
	}
	tech, err := CreateUser(ctx, database, admin.User, "tech", "tech12345", RoleTechnician)
	if err != nil {
		t.Fatal(err)
	}
	device, err := RegisterDevice(ctx, database, root, tech, NewDevice{
		CustomerName: "Ada", Brand: "HP", Model: "Laser", Note: "works",
	})
	if err != nil {
		t.Fatal(err)
	}

	samePhoto := uploadBytes(pngBytes())
	_, err = ChangeDeviceStatus(ctx, database, root, tech, device.ID, device.Status, "again", &samePhoto)
	var fields *FieldError
	if !errors.As(err, &fields) || !contains(fields.Invalid, "status") {
		t.Fatalf("same status: %v", err)
	}
	if countRows(t, database, `SELECT COUNT(*) FROM device_status_changes`) != 0 {
		t.Fatal("same status wrote history")
	}
	if countRows(t, database, `SELECT COUNT(*) FROM audit_events WHERE event = ?`, EventDeviceStatusChanged) != 0 {
		t.Fatal("same status wrote audit")
	}

	photo := uploadBytes(pngBytes())
	changed, err := ChangeDeviceStatus(ctx, database, root, tech, device.ID, StatusInspecting, "looked at it", &photo)
	if err != nil {
		t.Fatal(err)
	}
	if changed.Status != StatusInspecting || len(changed.History) != 1 || changed.History[0].Note != "looked at it" || !changed.History[0].HasPhoto {
		t.Fatalf("change %+v", changed.History)
	}
	if changed.History[0].From != StatusRegistered || changed.History[0].To != StatusInspecting || changed.History[0].ActorUsername != tech.Username {
		t.Fatalf("history row %+v", changed.History[0])
	}
	if countRows(t, database, `SELECT COUNT(*) FROM photos WHERE device_id = ?`, device.ID) != 0 {
		t.Fatal("status photo counted as a device photo")
	}
	full, media, err := OpenStatusPhoto(ctx, database, root, device.ID, changed.History[0].ID)
	if err != nil || media != "image/png" {
		t.Fatalf("photo %s %s %v", full, media, err)
	}
	if _, err := os.Stat(full); err != nil {
		t.Fatal(err)
	}

	var detail string
	if err := database.QueryRowContext(ctx, `
		SELECT detail FROM audit_events WHERE event = ? ORDER BY id DESC LIMIT 1
	`, EventDeviceStatusChanged).Scan(&detail); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(detail, "looked") || strings.Contains(detail, "devices") {
		t.Fatalf("audit stored the note or photo: %s", detail)
	}
	var parsed map[string]string
	if err := json.Unmarshal([]byte(detail), &parsed); err != nil {
		t.Fatal(err)
	}
	if parsed["from"] != StatusRegistered || parsed["to"] != StatusInspecting || parsed["uuid"] != device.ID {
		t.Fatalf("audit detail %s", detail)
	}

	kept, err := UpdateDevice(ctx, database, root, tech, device.ID, DeviceChange{
		Brand: "HP", Model: "Laser", Note: "works",
		CustomerName: "Ada", Parts: []PartChange{},
	})
	if err != nil {
		t.Fatal(err)
	}
	if kept.Status != StatusInspecting {
		t.Fatalf("correction changed status to %s", kept.Status)
	}

	cancelled, err := ChangeDeviceStatus(ctx, database, root, tech, device.ID, StatusCancelled, "", nil)
	if err != nil || cancelled.Status != StatusCancelled {
		t.Fatalf("cancel %+v %v", cancelled.Status, err)
	}

	again, err := ChangeDeviceStatus(ctx, database, root, tech, device.ID, StatusDelivered, "", nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(again.History) != 3 || again.History[0].To != StatusDelivered || again.History[1].To != StatusCancelled || again.History[2].To != StatusInspecting || again.History[0].HasPhoto {
		t.Fatalf("history order %+v", again.History)
	}

	matched, err := ListDevices(ctx, database, DeviceQuery{Status: StatusDelivered, Limit: 10})
	if err != nil || len(matched.Items) != 1 || matched.Items[0].ID != device.ID {
		t.Fatalf("status filter %+v %v", matched, err)
	}
	other, err := ListDevices(ctx, database, DeviceQuery{Status: StatusRegistered, Text: "Laser", Limit: 10})
	if err != nil || len(other.Items) != 0 {
		t.Fatalf("combined filter %+v %v", other, err)
	}
	if _, err := ListDevices(ctx, database, DeviceQuery{Status: "nope"}); !errors.Is(err, ErrInvalidStatus) {
		t.Fatalf("bad status filter: %v", err)
	}
}

func TestOptionalNotesAndAddress(t *testing.T) {
	ctx := context.Background()
	database := openTestDB(t)
	root := t.TempDir()
	if err := EnsureAdmin(ctx, database, "admin", "admin1234"); err != nil {
		t.Fatal(err)
	}
	admin, err := Login(ctx, database, "admin", "admin1234")
	if err != nil {
		t.Fatal(err)
	}

	device, err := RegisterDevice(ctx, database, root, admin.User, NewDevice{
		CustomerName:    "Ada",
		CustomerAddress: "Atatürk Cad. 1",
		Brand:           "HP",
		Model:           "Laser",
		Parts:           []NewPart{{Name: "Cable"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if device.Note != "" || len(device.Parts) != 1 || device.Parts[0].Note != "" || device.Customer.Address != "Atatürk Cad. 1" {
		t.Fatalf("optional fields %+v", device)
	}

	customers, err := ListCustomers(ctx, database, CustomerQuery{Text: "atatürk", Limit: 10})
	if err != nil || len(customers.Items) != 1 || customers.Items[0].Address != "Atatürk Cad. 1" {
		t.Fatalf("address search %+v %v", customers, err)
	}
	devices, err := ListDevices(ctx, database, DeviceQuery{Text: "cad", Limit: 10})
	if err != nil || len(devices.Items) != 1 || devices.Items[0].ID != device.ID {
		t.Fatalf("device address search %+v %v", devices, err)
	}

	updated, err := UpdateCustomer(ctx, database, admin.User, device.Customer.ID, "Ada", "", "", "  Yeni sokak  ")
	if err != nil || updated.Address != "Yeni sokak" {
		t.Fatalf("address update %+v %v", updated, err)
	}
	_, err = UpdateCustomer(ctx, database, admin.User, device.Customer.ID, "Ada", "", "", strings.Repeat("a", maxAddressLen+1))
	var fields *FieldError
	if !errors.As(err, &fields) || !contains(fields.TooLong, "address") {
		t.Fatalf("long address: %v", err)
	}

	cleared, err := UpdateDevice(ctx, database, root, admin.User, device.ID, DeviceChange{
		Brand: "HP", Model: "Laser", CustomerName: "Ada", CustomerAddress: "Yeni sokak",
		Parts: []PartChange{{ID: device.Parts[0].ID, Name: "Cable"}},
	})
	if err != nil || cleared.Note != "" || cleared.Parts[0].Note != "" {
		t.Fatalf("cleared notes %+v %v", cleared, err)
	}
}

func TestWidenDeviceStatus(t *testing.T) {
	ctx := context.Background()
	database, err := db.Open(ctx, filepath.Join(t.TempDir(), "old.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { database.Close() })
	if _, err := database.ExecContext(ctx, `
		CREATE TABLE customers (
			id INTEGER PRIMARY KEY,
			name TEXT NOT NULL,
			phone TEXT NOT NULL DEFAULT '',
			email TEXT NOT NULL DEFAULT '',
			name_fold TEXT NOT NULL,
			phone_fold TEXT NOT NULL,
			email_fold TEXT NOT NULL,
			created_at TEXT NOT NULL
		);
		CREATE TABLE devices (
			id TEXT PRIMARY KEY,
			customer_id INTEGER NOT NULL REFERENCES customers(id),
			name TEXT NOT NULL,
			brand TEXT NOT NULL,
			model TEXT NOT NULL,
			serial TEXT NOT NULL DEFAULT '',
			note TEXT NOT NULL,
			status TEXT NOT NULL DEFAULT 'registered' CHECK (status = 'registered'),
			name_fold TEXT NOT NULL,
			brand_fold TEXT NOT NULL,
			model_fold TEXT NOT NULL,
			serial_fold TEXT NOT NULL,
			created_at TEXT NOT NULL
		);
		INSERT INTO customers (id, name, phone, email, name_fold, phone_fold, email_fold, created_at)
		VALUES (1, 'Ada', '', '', 'ada', '', '', '2026-01-01T00:00:00.000Z');
		INSERT INTO devices (
			id, customer_id, name, brand, model, serial, note, status,
			name_fold, brand_fold, model_fold, serial_fold, created_at
		) VALUES ('dev', 1, 'Printer', 'HP', 'Laser', '', 'works', 'registered', 'printer', 'hp', 'laser', '', '2026-01-01T00:00:00.000Z');
		CREATE TABLE parts (
			id TEXT PRIMARY KEY,
			device_id TEXT NOT NULL REFERENCES devices(id),
			name TEXT NOT NULL,
			note TEXT NOT NULL,
			position INTEGER NOT NULL
		);
		INSERT INTO parts (id, device_id, name, note, position) VALUES ('part', 'dev', 'Cable', 'ok', 0);
	`); err != nil {
		t.Fatal(err)
	}
	if err := Migrate(ctx, database); err != nil {
		t.Fatal(err)
	}
	if _, err := database.ExecContext(ctx, `UPDATE devices SET status = ? WHERE id = 'dev'`, StatusInStorage); err != nil {
		t.Fatal(err)
	}
	var ddl string
	if err := database.QueryRowContext(ctx, `SELECT sql FROM sqlite_master WHERE type = 'table' AND name = 'devices'`).Scan(&ddl); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(ddl, "name_fold") || strings.Contains(ddl, "name TEXT") {
		t.Fatalf("device name column remained: %s", ddl)
	}
	var brand, model string
	if err := database.QueryRowContext(ctx, `SELECT brand, model FROM devices WHERE id = 'dev'`).Scan(&brand, &model); err != nil || brand != "HP" || model != "Laser" {
		t.Fatalf("device after migrate: %s %s %v", brand, model, err)
	}
	var name string
	if err := database.QueryRowContext(ctx, `SELECT name FROM parts WHERE device_id = 'dev'`).Scan(&name); err != nil || name != "Cable" {
		t.Fatalf("part after migrate: %s %v", name, err)
	}
	var address string
	if err := database.QueryRowContext(ctx, `SELECT address FROM customers WHERE id = 1`).Scan(&address); err != nil || address != "" {
		t.Fatalf("address after migrate: %q %v", address, err)
	}
}

func TestWidenCancelledStatus(t *testing.T) {
	ctx := context.Background()
	database, err := db.Open(ctx, filepath.Join(t.TempDir(), "prior.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { database.Close() })
	if _, err := database.ExecContext(ctx, `
		CREATE TABLE customers (
			id INTEGER PRIMARY KEY,
			name TEXT NOT NULL,
			phone TEXT NOT NULL DEFAULT '',
			email TEXT NOT NULL DEFAULT '',
			address TEXT NOT NULL DEFAULT '',
			name_fold TEXT NOT NULL,
			phone_fold TEXT NOT NULL,
			email_fold TEXT NOT NULL,
			address_fold TEXT NOT NULL DEFAULT '',
			created_at TEXT NOT NULL
		);
		CREATE TABLE devices (
			id TEXT PRIMARY KEY,
			customer_id INTEGER NOT NULL REFERENCES customers(id),
			brand TEXT NOT NULL,
			model TEXT NOT NULL,
			serial TEXT NOT NULL DEFAULT '',
			note TEXT NOT NULL,
			status TEXT NOT NULL DEFAULT 'registered' CHECK (status IN ('registered', 'inspecting', 'repair_warranty', 'repair_out_of_warranty', 'waiting_partner_part', 'waiting_own_part', 'final_testing', 'preparing_delivery', 'in_storage', 'left_storage', 'delivered')),
			brand_fold TEXT NOT NULL,
			model_fold TEXT NOT NULL,
			serial_fold TEXT NOT NULL,
			created_at TEXT NOT NULL
		);
		INSERT INTO customers (id, name, phone, email, address, name_fold, phone_fold, email_fold, address_fold, created_at)
		VALUES (1, 'Ada', '', '', '', 'ada', '', '', '', '2026-01-01T00:00:00.000Z');
		INSERT INTO devices (
			id, customer_id, brand, model, serial, note, status,
			brand_fold, model_fold, serial_fold, created_at
		) VALUES ('dev', 1, 'HP', 'Laser', '', 'works', 'delivered', 'hp', 'laser', '', '2026-01-01T00:00:00.000Z');
	`); err != nil {
		t.Fatal(err)
	}
	if _, err := database.ExecContext(ctx, `UPDATE devices SET status = 'cancelled' WHERE id = 'dev'`); err == nil {
		t.Fatal("old check accepted cancelled")
	}
	if err := Migrate(ctx, database); err != nil {
		t.Fatal(err)
	}
	if _, err := database.ExecContext(ctx, `UPDATE devices SET status = ? WHERE id = 'dev'`, StatusCancelled); err != nil {
		t.Fatal(err)
	}
}

func uploadBytes(body []byte) Upload {
	return Upload{Open: func() (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewReader(body)), nil
	}}
}

func pngBytes() []byte {
	return []byte{0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a}
}

func contains(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func countRows(t *testing.T, database *sql.DB, query string, args ...any) int {
	t.Helper()
	var count int
	if err := database.QueryRowContext(context.Background(), query, args...).Scan(&count); err != nil {
		t.Fatal(err)
	}
	return count
}
