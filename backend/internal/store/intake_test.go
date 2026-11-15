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
		Serial:        "SN-1",
		Note:          "works",
		Parts:         []NewPart{{Name: "Cable", Note: "ok", Photos: []int{0}}},
		Files:         []Upload{uploadBytes(pngBytes())},
	})
	if err != nil {
		t.Fatal(err)
	}
	if device.Status != StatusRegistered || device.Customer.Name != "Ada Lovelace" || device.Customer.Phone != "5551112233" || len(device.Parts) != 1 || len(device.Parts[0].Photos) != 1 {
		t.Fatalf("device %+v", device)
	}
	if _, err := os.Stat(filepath.Join(root, device.Parts[0].Photos[0].Path)); err != nil {
		t.Fatal(err)
	}
	if countRows(t, database, `SELECT COUNT(*) FROM audit_events WHERE event = ?`, EventCustomerCreated) != 1 {
		t.Fatal("customer audit")
	}
	if countRows(t, database, `SELECT COUNT(*) FROM audit_events WHERE event = ?`, EventServiceRegistered) != 1 {
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
		Brand: device.Brand, Model: device.Model, Serial: device.Serial, Note: device.Note,
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
		Brand: renamed.Brand, Model: renamed.Model, Serial: renamed.Serial, Note: renamed.Note,
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
		Brand: renamed.Brand, Model: renamed.Model, Serial: renamed.Serial, Note: "changed",
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
		Brand: kept.Brand, Model: kept.Model, Serial: kept.Serial, Note: kept.Note,
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
		CustomerName: "Big", CustomerPhone: "5550000001", Brand: "Big", Model: "Big", Serial: "SN-BIG", Note: "big",
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
		CustomerName: "Many", CustomerPhone: "5550000002", Brand: "Many", Model: "Many", Serial: "SN-MANY", Note: "many",
		Files: files, DevicePhotos: indexes,
	})
	if !errors.As(err, &fields) || fields.Photo != "limit" {
		t.Fatalf("limit: %v", err)
	}

	for _, model := range []string{"First", "Second", "Third"} {
		if _, err := RegisterDevice(ctx, database, root, owner, NewDevice{
			CustomerID: &kept.Customer.ID, Brand: "HP", Model: model, Serial: "SN-N", Note: "n",
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
		CustomerName: "Ada", CustomerPhone: "5552223344", Brand: "HP", Model: "Laser", Serial: "SN-STATUS", Note: "works",
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
	if countRows(t, database, `SELECT COUNT(*) FROM service_status_changes`) != 0 {
		t.Fatal("same status wrote history")
	}
	if countRows(t, database, `SELECT COUNT(*) FROM audit_events WHERE event = ?`, EventServiceStatusChanged) != 0 {
		t.Fatal("same status wrote audit")
	}

	photo := uploadBytes(pngBytes())
	changed, err := ChangeDeviceStatus(ctx, database, root, tech, device.ID, StatusRepairWarranty, "looked at it", &photo)
	if err != nil {
		t.Fatal(err)
	}
	if changed.Status != StatusRepairWarranty || len(changed.History) != 1 || changed.History[0].Note != "looked at it" || !changed.History[0].HasPhoto {
		t.Fatalf("change %+v", changed.History)
	}
	if changed.History[0].From != StatusRegistered || changed.History[0].To != StatusRepairWarranty || changed.History[0].ActorUsername != tech.Username {
		t.Fatalf("history row %+v", changed.History[0])
	}
	if countRows(t, database, `SELECT COUNT(*) FROM photos WHERE service_id = ?`, device.ID) != 0 {
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
	`, EventServiceStatusChanged).Scan(&detail); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(detail, "looked") || strings.Contains(detail, "devices") {
		t.Fatalf("audit stored the note or photo: %s", detail)
	}
	var parsed map[string]string
	if err := json.Unmarshal([]byte(detail), &parsed); err != nil {
		t.Fatal(err)
	}
	if parsed["from"] != StatusRegistered || parsed["to"] != StatusRepairWarranty || parsed["uuid"] != device.ID {
		t.Fatalf("audit detail %s", detail)
	}

	kept, err := UpdateDevice(ctx, database, root, tech, device.ID, DeviceChange{
		Brand: "HP", Model: "Laser", Serial: "SN-STATUS", Note: "works",
		CustomerName: "Ada", CustomerPhone: "5552223344", Parts: []PartChange{},
	})
	if err != nil {
		t.Fatal(err)
	}
	if kept.Status != StatusRepairWarranty {
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
	if len(again.History) != 3 || again.History[0].To != StatusDelivered || again.History[1].To != StatusCancelled || again.History[2].To != StatusRepairWarranty || again.History[0].HasPhoto {
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

func TestPhoneIdentifiesCustomer(t *testing.T) {
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

	first, err := RegisterDevice(ctx, database, root, admin.User, NewDevice{
		CustomerName: "Ada", CustomerPhone: "555 555 55 55", Brand: "HP", Model: "Laser", Serial: "SN-PHONE",
	})
	if err != nil || first.Customer.Phone != "5555555555" {
		t.Fatalf("stored phone %+v %v", first.Customer, err)
	}
	for _, phone := range []string{"5555555555", "05555555555", "0555555 55 55", "+90 555 555 55 55", "0090 555 555 55 55", "(0555) 555-55-55"} {
		_, err := RegisterDevice(ctx, database, root, admin.User, NewDevice{
			CustomerName: "Ada Other", CustomerPhone: phone, Brand: "HP", Model: "Other", Serial: "SN-OTHER",
		})
		var fields *FieldError
		if !errors.As(err, &fields) || !contains(fields.Taken, "phone") {
			t.Fatalf("%s: %v", phone, err)
		}
	}
	if countRows(t, database, `SELECT COUNT(*) FROM customers`) != 1 {
		t.Fatal("duplicate phone wrote a customer")
	}

	second, err := RegisterDevice(ctx, database, root, admin.User, NewDevice{
		CustomerName: "Ada", CustomerPhone: "555 111 22 33", Brand: "HP", Model: "Desk", Serial: "SN-DESK",
	})
	if err != nil || second.Customer.Phone != "5551112233" {
		t.Fatalf("second %+v %v", second.Customer, err)
	}
	_, err = RegisterDevice(ctx, database, root, admin.User, NewDevice{
		CustomerName: "Ada", Brand: "HP", Model: "Blank",
	})
	var fields *FieldError
	if !errors.As(err, &fields) || !contains(fields.Missing, "phone") {
		t.Fatalf("missing phone: %v", err)
	}
	if countRows(t, database, `SELECT COUNT(*) FROM customers`) != 2 {
		t.Fatal("missing phone wrote a customer")
	}

	listed, err := ListCustomers(ctx, database, CustomerQuery{Text: "ada", Limit: 10})
	if err != nil || len(listed.Items) != 2 {
		t.Fatalf("name search %+v %v", listed, err)
	}
	phones := map[string]int{}
	for _, item := range listed.Items {
		if item.Name != "Ada" {
			t.Fatal(item.Name)
		}
		phones[item.Phone]++
	}
	if phones["5555555555"] != 1 || phones["5551112233"] != 1 {
		t.Fatalf("phones %+v", listed.Items)
	}

	byPhone, err := ListCustomers(ctx, database, CustomerQuery{Text: "0555 555 55 55", Limit: 10})
	if err != nil || len(byPhone.Items) != 1 || byPhone.Items[0].ID != first.Customer.ID {
		t.Fatalf("phone search %+v %v", byPhone, err)
	}
	devices, err := ListDevices(ctx, database, DeviceQuery{Text: "+90 555 111 22 33", Limit: 10})
	if err != nil || len(devices.Items) != 1 || devices.Items[0].ID != second.ID {
		t.Fatalf("device phone search %+v %v", devices, err)
	}

	_, err = UpdateCustomer(ctx, database, admin.User, second.Customer.ID, "Ada", "05555555555", "", "")
	if !errors.As(err, &fields) || !contains(fields.Taken, "phone") {
		t.Fatalf("update taken: %v", err)
	}
	kept, err := GetCustomer(ctx, database, second.Customer.ID)
	if err != nil || kept.Phone != "5551112233" {
		t.Fatalf("phone changed %+v %v", kept, err)
	}
	_, err = UpdateCustomer(ctx, database, admin.User, first.Customer.ID, "Ada", "  ", "", "")
	if !errors.As(err, &fields) || !contains(fields.Missing, "phone") {
		t.Fatalf("cleared phone: %v", err)
	}
	kept, err = GetCustomer(ctx, database, first.Customer.ID)
	if err != nil || kept.Phone != "5555555555" {
		t.Fatalf("phone cleared %+v %v", kept, err)
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
		CustomerPhone:   "5553334455",
		CustomerAddress: "Atatürk Cad. 1",
		Brand:           "HP",
		Model:           "Laser",
		Serial:          "SN-NOTE",
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

	updated, err := UpdateCustomer(ctx, database, admin.User, device.Customer.ID, "Ada", "5553334455", "", "  Yeni sokak  ")
	if err != nil || updated.Address != "Yeni sokak" {
		t.Fatalf("address update %+v %v", updated, err)
	}
	_, err = UpdateCustomer(ctx, database, admin.User, device.Customer.ID, "Ada", "5553334455", "", strings.Repeat("a", maxAddressLen+1))
	var fields *FieldError
	if !errors.As(err, &fields) || !contains(fields.TooLong, "address") {
		t.Fatalf("long address: %v", err)
	}

	cleared, err := UpdateDevice(ctx, database, root, admin.User, device.ID, DeviceChange{
		Brand: "HP", Model: "Laser", Serial: "SN-NOTE", CustomerName: "Ada", CustomerPhone: "5553334455", CustomerAddress: "Yeni sokak",
		Parts: []PartChange{{ID: device.Parts[0].ID, Name: "Cable"}},
	})
	if err != nil || cleared.Note != "" || cleared.Parts[0].Note != "" {
		t.Fatalf("cleared notes %+v %v", cleared, err)
	}
}

func TestDeviceIdentity(t *testing.T) {
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
	first, err := RegisterDevice(ctx, database, root, admin.User, NewDevice{
		CustomerName: "Ada", CustomerPhone: "5554445566", Brand: "Acme", Model: "S1", Serial: " sn-9 ",
	})
	if err != nil {
		t.Fatal(err)
	}
	if first.Serial != "sn-9" || first.DeviceID == "" || first.DeviceID == first.ID {
		t.Fatalf("identity %+v", first)
	}
	_, err = RegisterDevice(ctx, database, root, admin.User, NewDevice{
		CustomerName: "Bea", CustomerPhone: "5554445567", Brand: "acme", Model: "s1", Serial: "SN-9",
	})
	var shop *InShopError
	if !errors.As(err, &shop) || shop.ServiceID != first.ID {
		t.Fatalf("open service: %v", err)
	}
	if countRows(t, database, `SELECT COUNT(*) FROM devices`) != 1 {
		t.Fatal("second device")
	}
	if _, err := ChangeDeviceStatus(ctx, database, root, admin.User, first.ID, StatusDelivered, "", nil); err != nil {
		t.Fatal(err)
	}
	second, err := RegisterDevice(ctx, database, root, admin.User, NewDevice{
		CustomerName: "Bea", CustomerPhone: "5554445567", Brand: "ACME", Model: "S1", Serial: "sn-9",
	})
	if err != nil {
		t.Fatal(err)
	}
	if second.DeviceID != first.DeviceID || second.Customer.Name != "Bea" || second.ID == first.ID {
		t.Fatalf("return %+v", second)
	}
	reloaded, err := GetDevice(ctx, database, first.ID)
	if err != nil || reloaded.Customer.Name != "Ada" {
		t.Fatalf("first customer %+v %v", reloaded.Customer, err)
	}
	_, err = RegisterDevice(ctx, database, root, admin.User, NewDevice{
		CustomerName: "Cara", CustomerPhone: "5554445568", Brand: "Northwind", Model: "S1", Serial: "sn-9",
	})
	if err != nil {
		t.Fatal(err)
	}
	if countRows(t, database, `SELECT COUNT(*) FROM devices`) != 2 {
		t.Fatal("other brand")
	}
	_, err = UpdateDevice(ctx, database, root, admin.User, second.ID, DeviceChange{
		Brand: "Acme", Model: "S1", Serial: "sn-9",
		CustomerName: "Bea", CustomerPhone: "5554445567",
	})
	if err != nil {
		t.Fatal(err)
	}
	moved, err := UpdateDevice(ctx, database, root, admin.User, second.ID, DeviceChange{
		Brand: "Northwind", Model: "S1", Serial: "sn-9",
		CustomerName: "Bea", CustomerPhone: "5554445567",
	})
	var fields *FieldError
	if !errors.As(err, &fields) || !contains(fields.Taken, "identity") || moved.ID != "" {
		t.Fatalf("collision %+v %v", moved, err)
	}
	stayed, err := GetDevice(ctx, database, second.ID)
	if err != nil || stayed.Brand != "Acme" || stayed.DeviceID != first.DeviceID {
		t.Fatalf("service moved %+v %v", stayed, err)
	}
	renamed, err := UpdateDevice(ctx, database, root, admin.User, first.ID, DeviceChange{
		Brand: "Acme", Model: "S1", Serial: "SN-10",
		CustomerName: "Ada", CustomerPhone: "5554445566",
	})
	if err != nil || renamed.Serial != "SN-10" || renamed.DeviceID != first.DeviceID {
		t.Fatalf("rename %+v %v", renamed, err)
	}
	other, err := GetDevice(ctx, database, second.ID)
	if err != nil || other.Serial != "SN-10" {
		t.Fatalf("shared identity %+v %v", other, err)
	}
	match, err := MatchDevice(ctx, database, "acme", "s1", "sn-10", VisitQuery{Limit: 10})
	if err != nil || !match.Matched || match.DeviceID != first.DeviceID || len(match.Items) != 2 || match.OpenServiceID != second.ID {
		t.Fatalf("match %+v %v", match, err)
	}
	if match.Items[0].ID != second.ID || match.Items[1].ID != first.ID {
		t.Fatalf("match order %+v", match.Items)
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
