package store

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestServiceComments(t *testing.T) {
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
	techUser, err := CreateUser(ctx, database, admin.User, "tech", "technician1", RoleTechnician)
	if err != nil {
		t.Fatal(err)
	}
	ownerUser, err := CreateUser(ctx, database, admin.User, "boss", "ownerpass1", RoleOwner)
	if err != nil {
		t.Fatal(err)
	}

	device, err := RegisterDevice(ctx, database, root, techUser, NewDevice{
		CustomerName:  "Ada",
		CustomerPhone: "555 111 22 33",
		Brand:         "HP",
		Model:         "Laser",
		Serial:        "SN-C1",
		Note:          "ok",
	})
	if err != nil {
		t.Fatal(err)
	}

	empty, err := GetDevice(ctx, database, device.ID)
	if err != nil || len(empty.Comments) != 0 {
		t.Fatalf("empty comments %+v %v", empty.Comments, err)
	}

	if _, err := AddServiceComment(ctx, database, techUser, device.ID, "   "); err == nil {
		t.Fatal("empty comment")
	}
	var fields *FieldError
	if _, err := AddServiceComment(ctx, database, techUser, device.ID, ""); !errors.As(err, &fields) || !contains(fields.Missing, "comment") {
		t.Fatalf("missing comment: %v", err)
	}
	long := strings.Repeat("a", maxNoteLen+1)
	if _, err := AddServiceComment(ctx, database, techUser, device.ID, long); !errors.As(err, &fields) || !contains(fields.TooLong, "comment") {
		t.Fatalf("too long: %v", err)
	}
	if _, err := AddServiceComment(ctx, database, techUser, "missing", "hello"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("missing service: %v", err)
	}

	auditsBefore := countRows(t, database, `SELECT COUNT(*) FROM audit_events`)

	first, err := AddServiceComment(ctx, database, ownerUser, device.ID, "  When will it be ready?  ")
	if err != nil {
		t.Fatal(err)
	}
	if first.Body != "When will it be ready?" || first.AuthorUsername != "boss" || first.AuthorID != ownerUser.ID || first.Edited {
		t.Fatalf("first %+v", first)
	}
	time.Sleep(2 * time.Millisecond)
	second, err := AddServiceComment(ctx, database, techUser, device.ID, "Tomorrow afternoon.")
	if err != nil {
		t.Fatal(err)
	}

	loaded, err := GetDevice(ctx, database, device.ID)
	if err != nil || len(loaded.Comments) != 2 {
		t.Fatalf("loaded %+v %v", loaded.Comments, err)
	}
	if loaded.Comments[0].ID != first.ID || loaded.Comments[1].ID != second.ID {
		t.Fatalf("order %+v", loaded.Comments)
	}

	same, err := UpdateServiceComment(ctx, database, ownerUser, device.ID, first.ID, "When will it be ready?")
	if err != nil {
		t.Fatal(err)
	}
	if same.Edited || same.UpdatedAt != first.UpdatedAt {
		t.Fatalf("unchanged edit %+v", same)
	}

	if _, err := UpdateServiceComment(ctx, database, techUser, device.ID, first.ID, "Not yours"); !errors.Is(err, ErrForbidden) {
		t.Fatalf("other author: %v", err)
	}

	edited, err := UpdateServiceComment(ctx, database, ownerUser, device.ID, first.ID, "Ready by Friday?")
	if err != nil {
		t.Fatal(err)
	}
	if !edited.Edited || edited.Body != "Ready by Friday?" || !edited.UpdatedAt.After(edited.CreatedAt) {
		t.Fatalf("edited %+v", edited)
	}

	if countRows(t, database, `SELECT COUNT(*) FROM audit_events`) != auditsBefore {
		t.Fatal("comment wrote audit")
	}

	delivered, err := ChangeDeviceStatus(ctx, database, root, techUser, device.ID, StatusDelivered, "", nil)
	if err != nil {
		t.Fatal(err)
	}
	after, err := AddServiceComment(ctx, database, admin.User, delivered.ID, "Customer collected.")
	if err != nil {
		t.Fatal(err)
	}
	final, err := GetDevice(ctx, database, delivered.ID)
	if err != nil || len(final.Comments) != 3 || final.Comments[2].ID != after.ID {
		t.Fatalf("after deliver %+v %v", final.Comments, err)
	}
}
