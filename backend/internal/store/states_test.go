package store

import (
	"context"
	"errors"
	"testing"
	"time"

	"techservice/internal/db"
)

func TestStateCatalog(t *testing.T) {
	ctx := context.Background()
	shop := openShop(t)
	items, err := ListStates(ctx, shop.db)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 5 || items[0].Code != StatusRegistered || items[4].Code != StatusDelivered {
		t.Fatalf("catalog %+v", items)
	}
	for _, item := range items {
		if item.ShowOnHome || !item.Active || !item.Fixed {
			t.Fatalf("seed %+v", item)
		}
	}
	if !items[0].Locked || !items[3].Locked || !items[4].Locked || items[1].Locked {
		t.Fatalf("locks %+v", items)
	}

	active := false
	if _, err := UpdateState(ctx, shop.db, shop.actor, StatusRegistered, StatePatch{Active: &active}); err == nil {
		t.Fatal("registered was deactivated")
	}
	show := true
	if _, err := UpdateState(ctx, shop.db, shop.actor, StatusDelivered, StatePatch{ShowOnHome: &show}); err == nil {
		t.Fatal("delivered was shown")
	}

	added, err := AddState(ctx, shop.db, shop.actor, StateInput{NameEn: "In storage", NameTr: "Depoda", ShowOnHome: true})
	if err != nil {
		t.Fatal(err)
	}
	if added.Code != "in_storage" || !added.Active || !added.ShowOnHome || added.Fixed {
		t.Fatalf("added %+v", added)
	}
	renamed, err := UpdateState(ctx, shop.db, shop.actor, added.Code, StatePatch{NameEn: strPtr("Stored"), NameTr: strPtr("Depoda")})
	if err != nil || renamed.Code != "in_storage" || renamed.NameEn != "Stored" {
		t.Fatalf("rename %+v %v", renamed, err)
	}
	if _, err := AddState(ctx, shop.db, shop.actor, StateInput{NameEn: "Stored", NameTr: "Başka"}); err == nil {
		t.Fatal("duplicate english name")
	}
	off := false
	if _, err := UpdateState(ctx, shop.db, shop.actor, StatusRepairWarranty, StatePatch{Active: &off}); err != nil {
		t.Fatal(err)
	}
	device := mustService(t, ctx, shop)
	if _, err := ChangeDeviceStatus(ctx, shop.db, shop.root, shop.actor, device, StatusRepairWarranty, "", nil); err == nil {
		t.Fatal("inactive state accepted")
	}

	moved, err := MoveState(ctx, shop.db, shop.actor, StatusDelivered, 4)
	if err != nil || moved.Position != 4 {
		t.Fatalf("position %+v %v", moved, err)
	}
	if _, err := MoveState(ctx, shop.db, shop.actor, StatusRegistered, 0); err == nil {
		t.Fatal("position 0 accepted")
	}
	same, err := MoveState(ctx, shop.db, shop.actor, StatusRegistered, 1)
	if err != nil || same.Position != 1 {
		t.Fatalf("same place %+v %v", same, err)
	}
	placed, err := MoveState(ctx, shop.db, shop.actor, StatusDelivered, 1)
	if err != nil || placed.Position != 1 {
		t.Fatalf("placed %+v %v", placed, err)
	}
	list, err := ListStates(ctx, shop.db)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{StatusDelivered, StatusRegistered, StatusRepairWarranty, StatusRepairOutOfWarranty, StatusCancelled, "in_storage"}
	if len(list) != len(want) {
		t.Fatalf("order length %d", len(list))
	}
	for i, code := range want {
		if list[i].Code != code || list[i].Position != i+1 {
			t.Fatalf("order %d %+v", i, list[i])
		}
	}
}

func TestLegacyStates(t *testing.T) {
	ctx := context.Background()
	database, err := db.Open(ctx, t.TempDir()+"/legacy.db")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { database.Close() })
	if _, err := database.ExecContext(ctx, `
		CREATE TABLE schema_migrations (
			version INTEGER PRIMARY KEY,
			applied_at TEXT NOT NULL
		)
	`); err != nil {
		t.Fatal(err)
	}
	if err := applyMigration(ctx, database, 1, false, migration1); err != nil {
		t.Fatal(err)
	}
	now := timestamp(time.Now())
	if _, err := database.ExecContext(ctx, `
		INSERT INTO customers (name, phone, email, address, name_fold, phone_fold, email_fold, address_fold, created_at)
		VALUES ('A', '5550001', '', '', 'a', '5550001', '', '', ?)
	`, now); err != nil {
		t.Fatal(err)
	}
	if _, err := database.ExecContext(ctx, `
		INSERT INTO devices (id, brand, model, serial, brand_fold, model_fold, serial_fold, created_at)
		VALUES ('dev', 'Acme', 'K1', 'S', 'acme', 'k1', 's', ?)
	`, now); err != nil {
		t.Fatal(err)
	}
	if _, err := database.ExecContext(ctx, `
		INSERT INTO services (id, device_id, customer_id, note, status, created_at)
		VALUES ('svc', 'dev', 1, '', 'inspecting', ?)
	`, now); err != nil {
		t.Fatal(err)
	}
	if _, err := database.ExecContext(ctx, `
		INSERT INTO service_status_changes (service_id, actor_username, from_status, to_status, note, created_at)
		VALUES ('svc', 'tech', 'left_storage', 'inspecting', '', ?)
	`, now); err != nil {
		t.Fatal(err)
	}
	if err := applyMigration(ctx, database, 2, true, migration2); err != nil {
		t.Fatal(err)
	}
	items, err := ListStates(ctx, database)
	if err != nil {
		t.Fatal(err)
	}
	got := map[string]ServiceState{}
	for _, item := range items {
		got[item.Code] = item
	}
	if !got[StatusInspecting].Active || got[StatusInspecting].Fixed || got[StatusInspecting].Position != 6 {
		t.Fatalf("inspecting %+v", got[StatusInspecting])
	}
	if got[StatusLeftStorage].Active || got[StatusLeftStorage].Position != 7 {
		t.Fatalf("left %+v", got[StatusLeftStorage])
	}
	if _, ok := got[StatusFinalTesting]; ok {
		t.Fatal("unused legacy state was written")
	}
	if !errors.Is(filterableStatus(ctx, database, StatusLeftStorage), ErrInvalidStatus) {
		t.Fatal("history-only state is filterable")
	}
	if err := filterableStatus(ctx, database, StatusInspecting); err != nil {
		t.Fatal(err)
	}
}

func strPtr(value string) *string {
	return &value
}

func mustService(t *testing.T, ctx context.Context, shop shopFix) string {
	t.Helper()
	device, err := RegisterDevice(ctx, shop.db, shop.root, shop.actor, NewDevice{
		CustomerName: "Ada", CustomerPhone: "5553334455", Brand: "Acme", Model: "K1", Serial: "STATE-" + time.Now().Format("150405.000"), Note: "n",
	})
	if err != nil {
		t.Fatal(err)
	}
	return device.ID
}
