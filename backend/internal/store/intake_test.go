package store

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"
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
	if !errors.As(err, &fields) || !contains(fields.Missing, "name") {
		t.Fatalf("missing name: %v", err)
	}
	if countRows(t, database, `SELECT COUNT(*) FROM customers`) != 0 {
		t.Fatal("invalid register wrote a customer")
	}

	_, err = RegisterDevice(ctx, database, root, tech, NewDevice{
		CustomerName: "Ada", Name: "Printer", Brand: "HP", Model: "Laser", Note: "ok", CustomerEmail: "nope",
	})
	if !errors.As(err, &fields) || !contains(fields.Invalid, "email") {
		t.Fatalf("email: %v", err)
	}

	device, err := RegisterDevice(ctx, database, root, tech, NewDevice{
		CustomerName:  "Ada Lovelace",
		CustomerPhone: "555",
		Name:          "Printer",
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
		Name: device.Name, Brand: device.Brand, Model: device.Model, Note: device.Note,
		CustomerName: "Ada King", CustomerPhone: "555",
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
		Name: renamed.Name, Brand: renamed.Brand, Model: renamed.Model, Note: renamed.Note,
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
		Name: renamed.Name, Brand: renamed.Brand, Model: renamed.Model, Note: "changed",
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
		Name: kept.Name, Brand: kept.Brand, Model: kept.Model, Note: kept.Note,
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
		CustomerName: "Big", Name: "Big", Brand: "Big", Model: "Big", Note: "big",
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
		CustomerName: "Many", Name: "Many", Brand: "Many", Model: "Many", Note: "many",
		Files: files, DevicePhotos: indexes,
	})
	if !errors.As(err, &fields) || fields.Photo != "limit" {
		t.Fatalf("limit: %v", err)
	}

	for i, name := range []string{"First", "Second", "Third"} {
		if _, err := RegisterDevice(ctx, database, root, owner, NewDevice{
			CustomerID: &kept.Customer.ID, Name: name, Brand: "HP", Model: "M", Note: "n",
		}); err != nil {
			t.Fatal(err)
		}
		_ = i
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
