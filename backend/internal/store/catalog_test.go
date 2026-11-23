package store

import (
	"context"
	"errors"
	"testing"

	"techservice/internal/catalog"
)

func TestCatalogMerge(t *testing.T) {
	ctx := context.Background()
	database := openTestDB(t)
	if err := EnsureAdmin(ctx, database, "admin", "admin1234"); err != nil {
		t.Fatal(err)
	}
	admin, err := Login(ctx, database, "admin", "admin1234")
	if err != nil {
		t.Fatal(err)
	}

	file := []catalog.Brand{{Brand: "Acme", Models: []string{"S1", "S2"}}}
	if err := MergeCatalog(ctx, database, file); err != nil {
		t.Fatal(err)
	}
	if err := MergeCatalog(ctx, database, file); err != nil {
		t.Fatal(err)
	}
	page, err := ListCatalogModels(ctx, database, CatalogListQuery{Limit: 10})
	if err != nil || len(page.Items) != 2 || page.Items[0].Brand != "Acme" || page.Items[0].Source != catalogFile || !page.Items[0].Active {
		t.Fatalf("merged %+v %v", page.Items, err)
	}
	if countRows(t, database, `SELECT COUNT(*) FROM audit_events WHERE event IN (?, ?)`, EventModelAdded, EventModelUpdated) != 0 {
		t.Fatal("startup merge wrote an audit row")
	}

	closed, err := SetCatalogActive(ctx, database, admin.User, page.Items[0].ID, false)
	if err != nil || closed.Active {
		t.Fatalf("close: %+v %v", closed, err)
	}
	if _, err := SetCatalogActive(ctx, database, admin.User, closed.ID, false); err != nil {
		t.Fatal(err)
	}
	if countRows(t, database, `SELECT COUNT(*) FROM audit_events WHERE event = ?`, EventModelUpdated) != 1 {
		t.Fatal("unchanged active flag wrote an audit row")
	}
	if err := MergeCatalog(ctx, database, file); err != nil {
		t.Fatal(err)
	}
	again, err := loadCatalogModel(ctx, database, closed.ID)
	if err != nil || again.Active || again.Source != catalogFile || again.Brand != "Acme" {
		t.Fatalf("file left an inactive catalog row: %+v %v", again, err)
	}

	added, err := CreateCatalogModel(ctx, database, admin.User, " North ", "Mini 2")
	if err != nil || added.Brand != "North" || added.Model != "Mini 2" || added.Source != catalogAdded || !added.Active {
		t.Fatalf("added %+v %v", added, err)
	}
	if _, err := CreateCatalogModel(ctx, database, admin.User, "north", "mini 2"); !fieldTaken(err, "catalog") {
		t.Fatalf("duplicate: %v", err)
	}
	if _, err := CreateCatalogModel(ctx, database, admin.User, " ", "Mini"); !fieldMissing(err, "brand") {
		t.Fatalf("blank brand: %v", err)
	}

	if err := MergeCatalog(ctx, database, []catalog.Brand{{Brand: "Acme", Models: []string{"S2"}}}); err != nil {
		t.Fatal(err)
	}
	left, err := loadCatalogModel(ctx, database, page.Items[1].ID)
	if err != nil || !left.Active || left.Model != "S2" {
		t.Fatalf("listed model: %+v %v", left, err)
	}
	dropped, err := loadCatalogModel(ctx, database, closed.ID)
	if err != nil || dropped.Active {
		t.Fatalf("removed catalog model stayed active: %+v %v", dropped, err)
	}
	kept, err := loadCatalogModel(ctx, database, added.ID)
	if err != nil || !kept.Active || kept.Source != catalogAdded {
		t.Fatalf("added model changed: %+v %v", kept, err)
	}
	if _, err := SetCatalogActive(ctx, database, admin.User, dropped.ID, true); err != nil {
		t.Fatal(err)
	}
	if err := MergeCatalog(ctx, database, []catalog.Brand{{Brand: "Acme", Models: []string{"S2"}}}); err != nil {
		t.Fatal(err)
	}
	reopened, err := loadCatalogModel(ctx, database, dropped.ID)
	if err != nil || reopened.Active || reopened.Source != catalogFile {
		t.Fatalf("reopened catalog model survived the next start: %+v %v", reopened, err)
	}
	if err := MergeCatalog(ctx, database, file); err != nil {
		t.Fatal(err)
	}
	back, err := loadCatalogModel(ctx, database, dropped.ID)
	if err != nil || back.Active {
		t.Fatalf("returning file pair was forced active: %+v %v", back, err)
	}

	found, err := ListCatalogModels(ctx, database, CatalogListQuery{Text: "mini", Limit: 10})
	if err != nil || len(found.Items) != 1 || found.Items[0].ID != added.ID {
		t.Fatalf("search %+v %v", found.Items, err)
	}
	brands, err := SuggestCatalogBrands(ctx, database, "", true)
	if err != nil || len(brands) != 2 || brands[0] != "Acme" || brands[1] != "North" {
		t.Fatalf("active brands %#v %v", brands, err)
	}
	models, err := SuggestCatalogModels(ctx, database, "acme", "", true)
	if err != nil || len(models) != 1 || models[0].Model != "S2" {
		t.Fatalf("active models %#v %v", models, err)
	}
	all, err := SuggestCatalogModels(ctx, database, "acme", "", false)
	if err != nil || len(all) != 2 {
		t.Fatalf("all models %#v %v", all, err)
	}
}

func TestCatalogFromService(t *testing.T) {
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

	if _, err := RegisterDevice(ctx, database, root, tech, NewDevice{
		CustomerName: "Ada", CustomerPhone: "5551112233", Brand: " ", Model: "Laser", Serial: "SN-1",
	}); !fieldMissing(err, "brand") {
		t.Fatal(err)
	}
	if countRows(t, database, `SELECT COUNT(*) FROM catalog_models`) != 0 {
		t.Fatal("refused registration stored a model")
	}

	device, err := RegisterDevice(ctx, database, root, tech, NewDevice{
		CustomerName: "Ada", CustomerPhone: "5551112233", Brand: "HP", Model: "Laser", Serial: "SN-1",
	})
	if err != nil {
		t.Fatal(err)
	}
	rows, err := database.QueryContext(ctx, `
		SELECT event FROM audit_events
		WHERE event IN (?, ?, ?)
		ORDER BY id
	`, EventCustomerCreated, EventModelAdded, EventServiceRegistered)
	if err != nil {
		t.Fatal(err)
	}
	var events []string
	for rows.Next() {
		var event string
		if err := rows.Scan(&event); err != nil {
			t.Fatal(err)
		}
		events = append(events, event)
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	rows.Close()
	if len(events) != 3 || events[0] != EventCustomerCreated || events[1] != EventModelAdded || events[2] != EventServiceRegistered {
		t.Fatalf("audit order %#v", events)
	}

	if _, err := RegisterDevice(ctx, database, root, tech, NewDevice{
		CustomerID: &device.Customer.ID, Brand: "hp", Model: "laser", Serial: "SN-2",
	}); err != nil {
		t.Fatal(err)
	}
	if countRows(t, database, `SELECT COUNT(*) FROM catalog_models`) != 1 {
		t.Fatal("case variant stored a second model")
	}
	if countRows(t, database, `SELECT COUNT(*) FROM audit_events WHERE event = ?`, EventModelAdded) != 1 {
		t.Fatal("matched model wrote another audit row")
	}
	stored, err := findCatalogModel(ctx, database, "hp", "laser")
	if err != nil || !stored.Active || stored.Brand != "HP" {
		t.Fatalf("stored %+v %v", stored, err)
	}
	if _, err := SetCatalogActive(ctx, database, tech, stored.ID, false); err != nil {
		t.Fatal(err)
	}
	if _, err := RegisterDevice(ctx, database, root, tech, NewDevice{
		CustomerID: &device.Customer.ID, Brand: "HP", Model: "Laser", Serial: "SN-3",
	}); err != nil {
		t.Fatal(err)
	}
	inactive, err := loadCatalogModel(ctx, database, stored.ID)
	if err != nil || inactive.Active {
		t.Fatalf("registration reactivated %+v %v", inactive, err)
	}

	other, err := RegisterDevice(ctx, database, root, tech, NewDevice{
		CustomerID: &device.Customer.ID, Brand: "Dell", Model: "X", Serial: "SN-9",
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := database.ExecContext(ctx, `DELETE FROM catalog_models WHERE brand_fold = ? AND model_fold = ?`, fold("Dell"), fold("X")); err != nil {
		t.Fatal(err)
	}
	_, err = UpdateDevice(ctx, database, root, tech, device.ID, DeviceChange{
		Brand: "Dell", Model: "X", Serial: "SN-9", Note: device.Note,
		CustomerName: device.Customer.Name, CustomerPhone: "5551112233",
	})
	var fields *FieldError
	if !errors.As(err, &fields) || !contains(fields.Taken, "identity") {
		t.Fatalf("identity: %v", err)
	}
	if countRows(t, database, `SELECT COUNT(*) FROM catalog_models WHERE brand_fold = ?`, fold("Dell")) != 0 {
		t.Fatal("refused correction stored a model")
	}

	corrected, err := UpdateDevice(ctx, database, root, tech, other.ID, DeviceChange{
		Brand: "Dell", Model: "Y", Serial: "SN-9", Note: other.Note,
		CustomerName: device.Customer.Name, CustomerPhone: "5551112233",
	})
	if err != nil || corrected.Model != "Y" {
		t.Fatalf("correction %+v %v", corrected.Model, err)
	}
	fresh, err := findCatalogModel(ctx, database, "Dell", "Y")
	if err != nil || !fresh.Active || fresh.Source != catalogAdded {
		t.Fatalf("correction model %+v %v", fresh, err)
	}
}
