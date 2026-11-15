package store

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"strings"
	"testing"
)

func TestSpareParts(t *testing.T) {
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
	actor := tech

	qty := func(n int) *float64 {
		value := float64(n)
		return &value
	}
	fraction := 1.5

	if _, err := CreateSparePart(ctx, database, actor, "  ", "A52", qty(1)); !fieldMissing(err, "brand") {
		t.Fatalf("brand: %v", err)
	}
	if _, err := CreateSparePart(ctx, database, actor, "Samsung", "A52", nil); !fieldMissing(err, "quantity") {
		t.Fatalf("quantity: %v", err)
	}
	if _, err := CreateSparePart(ctx, database, actor, "Samsung", "A52", &fraction); !fieldInvalid(err, "quantity") {
		t.Fatalf("fraction: %v", err)
	}
	if countRows(t, database, `SELECT COUNT(*) FROM spare_parts`) != 0 {
		t.Fatal("invalid create wrote a part")
	}

	created, err := CreateSparePart(ctx, database, actor, " Samsung ", "A52", qty(0))
	if err != nil || created.Existed || created.Added || created.OnHand != 0 || created.Brand != "Samsung" {
		t.Fatalf("empty create %+v %v", created, err)
	}
	if countRows(t, database, `SELECT COUNT(*) FROM spare_part_moves WHERE part_id = ?`, created.ID) != 0 {
		t.Fatal("zero quantity wrote a receipt")
	}
	if countRows(t, database, `SELECT COUNT(*) FROM audit_events WHERE event = ?`, EventSparePartAdded) != 1 {
		t.Fatal("create audit missing")
	}

	again, err := CreateSparePart(ctx, database, actor, "samsung", "a52", qty(0))
	if err != nil || !again.Existed || again.Added || again.ID != created.ID || again.OnHand != 0 {
		t.Fatalf("duplicate zero %+v %v", again, err)
	}
	if countRows(t, database, `SELECT COUNT(*) FROM spare_parts`) != 1 {
		t.Fatal("duplicate created a second part")
	}
	if countRows(t, database, `SELECT COUNT(*) FROM audit_events WHERE event = ?`, EventSparePartAdded) != 1 {
		t.Fatal("duplicate zero wrote an audit row")
	}

	added, err := CreateSparePart(ctx, database, actor, "SAMSUNG", "A52", qty(4))
	if err != nil || !added.Existed || !added.Added || added.ID != created.ID || added.OnHand != 4 || added.Brand != "Samsung" {
		t.Fatalf("duplicate add %+v %v", added, err)
	}
	if countRows(t, database, `SELECT COUNT(*) FROM spare_part_moves WHERE part_id = ? AND kind = ?`, created.ID, moveReceipt) != 1 {
		t.Fatal("duplicate add receipt")
	}

	screen, err := CreateSparePart(ctx, database, actor, "Samsung", "S21", qty(2))
	if err != nil || screen.OnHand != 2 {
		t.Fatalf("second part %+v %v", screen, err)
	}
	page, err := ListSpareParts(ctx, database, SpareListQuery{Text: "sam", Limit: 1})
	if err != nil || len(page.Items) != 1 || !page.HasMore || page.Items[0].Model != "A52" {
		t.Fatalf("page %+v %v", page, err)
	}
	next, err := ListSpareParts(ctx, database, SpareListQuery{
		Text: "sam", Limit: 1, HasAfter: true, AfterBrand: page.Items[0].Brand, AfterModel: page.Items[0].Model, AfterID: page.Items[0].ID,
	})
	if err != nil || len(next.Items) != 1 || next.HasMore || next.Items[0].ID != screen.ID {
		t.Fatalf("next page %+v %v", next, err)
	}
	brands, err := SuggestSpareBrands(ctx, database, "sam")
	if err != nil || len(brands) != 1 || brands[0] != "Samsung" {
		t.Fatalf("brands %+v %v", brands, err)
	}
	models, err := SuggestSpareModels(ctx, database, "samsung", "s")
	if err != nil || len(models) != 1 || models[0] != "S21" {
		t.Fatalf("models %+v %v", models, err)
	}
	if names, err := SuggestSpareModels(ctx, database, "Other", ""); err != nil || len(names) != 0 {
		t.Fatalf("unknown brand %+v %v", names, err)
	}

	same, err := UpdateSparePart(ctx, database, actor, created.ID, "Samsung", "A52")
	if err != nil || same.Brand != "Samsung" {
		t.Fatalf("same update %+v %v", same, err)
	}
	if countRows(t, database, `SELECT COUNT(*) FROM audit_events WHERE event = ?`, EventSparePartUpdated) != 0 {
		t.Fatal("unchanged correction wrote audit")
	}
	if _, err := UpdateSparePart(ctx, database, actor, created.ID, "Samsung", "S21"); !fieldTaken(err, "spare") {
		t.Fatalf("taken: %v", err)
	}
	renamed, err := UpdateSparePart(ctx, database, actor, created.ID, "Samsung", "A52b")
	if err != nil || renamed.Model != "A52b" || renamed.OnHand != 4 {
		t.Fatalf("rename %+v %v", renamed, err)
	}

	part, err := WriteOffSpareStock(ctx, database, actor, renamed.ID, qty(1), "cracked")
	if err != nil || part.OnHand != 3 {
		t.Fatalf("write off %+v %v", part, err)
	}
	if _, err := WriteOffSpareStock(ctx, database, actor, renamed.ID, qty(4), ""); !stockIs(err, StockShort) {
		t.Fatalf("short: %v", err)
	}
	if got := onHand(t, database, renamed.ID); got != 3 {
		t.Fatalf("short wrote %d", got)
	}
	var detail string
	if err := database.QueryRowContext(ctx, `SELECT detail FROM audit_events WHERE event = ? ORDER BY id DESC LIMIT 1`, EventStockWrittenOff).Scan(&detail); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(detail, "cracked") {
		t.Fatalf("audit contains the note %s", detail)
	}

	service := registerSpareService(t, ctx, database, root, actor, "HP", "Laser", "SN-1")
	onHandLeft, err := UseSparePart(ctx, database, actor, service.ID, renamed.ID, qty(2))
	if err != nil || onHandLeft != 1 {
		t.Fatalf("use %d %v", onHandLeft, err)
	}
	if _, err := ChangeDeviceStatus(ctx, database, root, actor, service.ID, StatusDelivered, "", nil); err != nil {
		t.Fatal(err)
	}
	if _, err := UseSparePart(ctx, database, actor, service.ID, renamed.ID, qty(1)); err != nil {
		t.Fatalf("use delivered: %v", err)
	}
	if _, err := UseSparePart(ctx, database, actor, service.ID, renamed.ID, qty(1)); !stockIs(err, StockShort) {
		t.Fatalf("empty use: %v", err)
	}
	fresh, err := GetDevice(ctx, database, service.ID)
	if err != nil || fresh.Status != StatusDelivered {
		t.Fatalf("status changed %+v %v", fresh.Status, err)
	}
	if _, err := UseSparePart(ctx, database, actor, service.ID, screen.ID, qty(1)); err != nil {
		t.Fatal(err)
	}
	full, err := AddSpareStock(ctx, database, actor, screen.ID, qty(99998), "delivery")
	if err != nil || full.OnHand != 99999 {
		t.Fatalf("fill %+v %v", full, err)
	}
	if _, err := ReturnSparePart(ctx, database, actor, service.ID, screen.ID, qty(1), "wrong"); !stockIs(err, StockCap) {
		t.Fatalf("return cap: %v", err)
	}
	if got := onHand(t, database, screen.ID); got != 99999 {
		t.Fatalf("return cap wrote %d", got)
	}
	if err := database.QueryRowContext(ctx, `SELECT detail FROM audit_events WHERE event = ? ORDER BY id DESC LIMIT 1`, EventStockAdded).Scan(&detail); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(detail, "delivery") {
		t.Fatalf("audit contains the note %s", detail)
	}

	uses, err := ListSpareUses(ctx, database, renamed.ID, SpareMoveQuery{})
	if err != nil || len(uses.Items) != 2 || uses.Items[0].Kind != moveUse || uses.Items[0].Status != StatusDelivered || uses.Items[0].Serial != "SN-1" {
		t.Fatalf("uses %+v %v", uses.Items, err)
	}
	if uses.Items[0].Model != "" || uses.Items[1].Status != StatusRegistered {
		t.Fatalf("use order %+v", uses.Items)
	}
	stock, err := ListSpareStock(ctx, database, renamed.ID, SpareMoveQuery{Limit: 1})
	if err != nil || len(stock.Items) != 1 || !stock.HasMore || stock.Items[0].Kind != moveWriteOff || stock.Items[0].Note != "cracked" {
		t.Fatalf("stock %+v %v", stock, err)
	}
	serviceUses, err := ListServiceSpares(ctx, database, service.ID, SpareMoveQuery{})
	if err != nil || len(serviceUses.Items) != 3 || serviceUses.Items[0].Brand != "Samsung" || serviceUses.Items[0].Model != "S21" {
		t.Fatalf("service uses %+v %v", serviceUses.Items, err)
	}

	returned, err := ReturnSparePart(ctx, database, actor, service.ID, renamed.ID, qty(3), "counted twice")
	if err != nil || returned != 3 {
		t.Fatalf("return %d %v", returned, err)
	}
	if _, err := ReturnSparePart(ctx, database, actor, service.ID, renamed.ID, qty(1), ""); !stockIs(err, ReturnShort) {
		t.Fatalf("return short: %v", err)
	}
	choices, err := ListReturnableSpares(ctx, database, service.ID, SpareListQuery{})
	if err != nil || len(choices.Items) != 1 || choices.Items[0].ID != screen.ID || choices.Items[0].Returnable != 1 {
		t.Fatalf("returnable %+v %v", choices, err)
	}
	listed, err := GetSparePart(ctx, database, renamed.ID)
	if err != nil || listed.Model != "A52b" {
		t.Fatalf("current name %+v %v", listed, err)
	}
	uses, err = ListSpareUses(ctx, database, renamed.ID, SpareMoveQuery{})
	if err != nil || len(uses.Items) != 3 || uses.Items[0].Kind != moveReturn || uses.Items[0].Note != "counted twice" {
		t.Fatalf("return row %+v %v", uses.Items, err)
	}
}

func registerSpareService(t *testing.T, ctx context.Context, database *sql.DB, root string, actor User, brand string, model string, serial string) Device {
	t.Helper()
	device, err := RegisterDevice(ctx, database, root, actor, NewDevice{
		CustomerName:  "Ada Lovelace",
		CustomerPhone: "5551112233",
		Brand:         brand,
		Model:         model,
		Serial:        serial,
		Note:          "ok",
	})
	if err != nil {
		t.Fatal(err)
	}
	return device
}

func onHand(t *testing.T, database *sql.DB, id string) int {
	t.Helper()
	var qty int
	if err := database.QueryRow(`SELECT on_hand FROM spare_parts WHERE id = ?`, id).Scan(&qty); err != nil {
		t.Fatal(err)
	}
	return qty
}

func fieldMissing(err error, field string) bool {
	var fields *FieldError
	return errors.As(err, &fields) && contains(fields.Missing, field)
}

func fieldInvalid(err error, field string) bool {
	var fields *FieldError
	return errors.As(err, &fields) && contains(fields.Invalid, field)
}

func fieldTaken(err error, field string) bool {
	var fields *FieldError
	return errors.As(err, &fields) && contains(fields.Taken, field)
}

func stockIs(err error, kind string) bool {
	var stock *StockError
	return errors.As(err, &stock) && stock.Kind == kind
}

func TestSpareQuantityEdges(t *testing.T) {
	ctx := context.Background()
	database := openTestDB(t)
	if err := EnsureAdmin(ctx, database, "admin", "admin1234"); err != nil {
		t.Fatal(err)
	}
	admin, err := Login(ctx, database, "admin", "admin1234")
	if err != nil {
		t.Fatal(err)
	}
	over := float64(maxStock + 1)
	if _, err := CreateSparePart(ctx, database, admin.User, "Acme", "One", &over); !fieldInvalid(err, "quantity") {
		t.Fatalf("over: %v", err)
	}
	value := float64(maxStock)
	part, err := CreateSparePart(ctx, database, admin.User, "Acme", "One", &value)
	if err != nil || part.OnHand != maxStock {
		t.Fatalf("max %+v %v", part, err)
	}
	if _, err := UseSparePart(ctx, database, admin.User, "missing", part.ID, floatPtr(1)); !errors.Is(err, ErrNotFound) {
		t.Fatalf("missing service: %v", err)
	}
	if _, err := GetSparePart(ctx, database, "missing"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("missing part: %v", err)
	}
	long := strings.Repeat("a", maxNameLen+1)
	if _, err := CreateSparePart(ctx, database, admin.User, long, "One", floatPtr(1)); !fieldTooLong(err, "brand") {
		t.Fatalf("long brand: %v", err)
	}
	note := strings.Repeat("n", maxNoteLen+1)
	if _, err := AddSpareStock(ctx, database, admin.User, part.ID, floatPtr(1), note); !fieldTooLong(err, "spareNote") {
		t.Fatalf("long note: %v", err)
	}
	if got := strconv.Itoa(onHand(t, database, part.ID)); got != strconv.Itoa(maxStock) {
		t.Fatalf("note wrote stock %s", got)
	}
}

func fieldTooLong(err error, field string) bool {
	var fields *FieldError
	return errors.As(err, &fields) && contains(fields.TooLong, field)
}

func floatPtr(n int) *float64 {
	value := float64(n)
	return &value
}
