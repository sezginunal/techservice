package store

import (
	"context"
	"database/sql"
	"testing"
	"time"
)

func TestSummaryEmpty(t *testing.T) {
	ctx := context.Background()
	shop := openShop(t)
	summary, err := LoadSummary(ctx, shop.db, summaryNow())
	if err != nil {
		t.Fatal(err)
	}
	if summary.Registered7 != 0 || summary.Registered30 != 0 || summary.InShop != 0 || summary.Delivered != 0 {
		t.Fatalf("counts %+v", summary)
	}
	if summary.Idle7 != 0 || summary.WaitingPart != 0 || summary.Ready != 0 || summary.Other != 0 || summary.UnchangedMore {
		t.Fatalf("shop counts %+v", summary)
	}
	if len(summary.Days) != homeDays || summary.Days[0].Date != "2026-10-05" || summary.Days[29].Date != "2026-11-03" {
		t.Fatalf("days %+v", summary.Days)
	}
	for _, day := range summary.Days {
		if day.Count != 0 {
			t.Fatalf("day %s count %d", day.Date, day.Count)
		}
	}
	if len(summary.ByStatus) != len(deviceStatusList()) || summary.ByStatus[0].Status != StatusRegistered || summary.ByStatus[len(summary.ByStatus)-1].Status != StatusCancelled {
		t.Fatalf("status order %+v", summary.ByStatus)
	}
	for _, item := range summary.ByStatus {
		if item.Count != 0 {
			t.Fatalf("status %s count %d", item.Status, item.Count)
		}
	}
	if len(summary.Models) != 0 || len(summary.Unchanged) != 0 || len(summary.Latest) != 0 {
		t.Fatalf("lists %+v", summary)
	}
}

func TestSummaryRegistrations(t *testing.T) {
	ctx := context.Background()
	shop := openShop(t)
	now := summaryNow()
	loc := istanbul
	var customer int64
	var newest []string
	add := func(brand, model string, at time.Time) string {
		t.Helper()
		var id *int64
		if customer != 0 {
			id = &customer
		}
		device := registerAt(t, ctx, shop, id, brand, model, at)
		if customer == 0 {
			customer = device.Customer.ID
		}
		deliver(t, ctx, shop, device.ID)
		return device.ID
	}

	for i := 0; i < 5; i++ {
		newest = append(newest, add("Acme", "K1", now.Add(-time.Duration(i)*time.Minute)))
	}
	for i := 0; i < 4; i++ {
		add("Acme", "Desk", time.Date(2026, 10, 28, 0, 30, 0, 0, loc))
	}
	for i := 0; i < 3; i++ {
		add("Northwind", "A1", time.Date(2026, 10, 27, 12, 0, 0, 0, loc))
	}
	for i := 0; i < 2; i++ {
		add("Helios", "R4", now.Add(-2*time.Hour))
	}
	for _, model := range []string{"A", "B", "C", "D", "E", "F"} {
		add("HP", model, now.Add(-2*time.Hour))
	}
	add("Boundary", "In", time.Date(2026, 10, 4, 21, 0, 0, 0, time.UTC))
	add("Anycubic", "Kobra", time.Date(2026, 10, 4, 20, 59, 59, 0, time.UTC))

	summary, err := LoadSummary(ctx, shop.db, now)
	if err != nil {
		t.Fatal(err)
	}
	if summary.Registered7 != 17 || summary.Registered30 != 21 || summary.Delivered != 22 || summary.InShop != 0 {
		t.Fatalf("counts 7=%d 30=%d delivered=%d inShop=%d", summary.Registered7, summary.Registered30, summary.Delivered, summary.InShop)
	}
	if dayCount(summary.Days, "2026-11-03") != 13 || dayCount(summary.Days, "2026-10-28") != 4 || dayCount(summary.Days, "2026-10-27") != 3 || dayCount(summary.Days, "2026-10-05") != 1 || dayCount(summary.Days, "2026-10-04") != -1 {
		t.Fatalf("days %+v", summary.Days)
	}
	if summary.Idle7 != 0 || len(summary.Unchanged) != 0 {
		t.Fatalf("delivered devices are idle %+v", summary.Unchanged)
	}
	if len(summary.Models) != 8 || summary.Other != 3 {
		t.Fatalf("models %d other %d %+v", len(summary.Models), summary.Other, summary.Models)
	}
	wantModels := []ModelCount{
		{"Acme", "K1", 5},
		{"Acme", "Desk", 4},
		{"Northwind", "A1", 3},
		{"Helios", "R4", 2},
		{"Boundary", "In", 1},
		{"HP", "A", 1},
		{"HP", "B", 1},
		{"HP", "C", 1},
	}
	for i, want := range wantModels {
		got := summary.Models[i]
		if got.Brand != want.Brand || got.Model != want.Model || got.Count != want.Count {
			t.Fatalf("model %d %+v", i, got)
		}
	}
	for _, item := range summary.Models {
		if item.Brand == "Anycubic" {
			t.Fatal("outside model counted")
		}
	}
	if len(summary.Latest) != 5 {
		t.Fatalf("latest %+v", summary.Latest)
	}
	for i, id := range newest {
		if summary.Latest[i].ID != id || summary.Latest[i].Brand != "Acme" || summary.Latest[i].Model != "K1" || summary.Latest[i].Status != StatusDelivered {
			t.Fatalf("latest %d %+v", i, summary.Latest[i])
		}
	}
}

func TestSummaryShop(t *testing.T) {
	ctx := context.Background()
	shop := openShop(t)
	now := summaryNow()
	var customer int64
	open := func(brand, model, status string) {
		t.Helper()
		var id *int64
		if customer != 0 {
			id = &customer
		}
		device := registerAt(t, ctx, shop, id, brand, model, now)
		if customer == 0 {
			customer = device.Customer.ID
		}
		if status != StatusRegistered {
			setStatusAt(t, ctx, shop, device.ID, status, now)
		}
	}
	open("Shop", "Registered", StatusRegistered)
	open("Shop", "Inspecting", StatusInspecting)
	open("Shop", "Partner", StatusWaitingPartnerPart)
	open("Shop", "Own", StatusWaitingOwnPart)
	open("Shop", "Ready", StatusPreparingDelivery)
	open("Shop", "Done", StatusDelivered)
	open("Shop", "Stored", StatusInStorage)
	open("Shop", "Cancelled", StatusCancelled)

	summary, err := LoadSummary(ctx, shop.db, now)
	if err != nil {
		t.Fatal(err)
	}
	if summary.InShop != 7 || summary.Delivered != 1 || summary.WaitingPart != 2 || summary.Ready != 1 || summary.Idle7 != 0 || len(summary.Unchanged) != 0 {
		t.Fatalf("shop %+v unchanged %d", summary, len(summary.Unchanged))
	}
	got := map[string]int{}
	for _, item := range summary.ByStatus {
		got[item.Status] = item.Count
	}
	if got[StatusRegistered] != 1 || got[StatusInspecting] != 1 || got[StatusWaitingPartnerPart] != 1 || got[StatusWaitingOwnPart] != 1 || got[StatusPreparingDelivery] != 1 || got[StatusDelivered] != 1 || got[StatusInStorage] != 1 || got[StatusCancelled] != 1 || got[StatusFinalTesting] != 0 {
		t.Fatalf("by status %+v", summary.ByStatus)
	}
	if summary.ByStatus[0].Status != StatusRegistered || summary.ByStatus[len(summary.ByStatus)-1].Status != StatusCancelled {
		t.Fatalf("order %+v", summary.ByStatus)
	}
}

func TestSummaryIdleClock(t *testing.T) {
	ctx := context.Background()
	shop := openShop(t)
	now := summaryNow()
	var customer int64
	add := func(at time.Time) Device {
		t.Helper()
		var id *int64
		if customer != 0 {
			id = &customer
		}
		device := registerAt(t, ctx, shop, id, "Idle", "Clock", at)
		if customer == 0 {
			customer = device.Customer.ID
		}
		return device
	}

	oneDay := add(now.Add(-fullDay))
	add(now.Add(-fullDay + time.Millisecond))
	seven := add(now.Add(-7 * fullDay))
	almostSeven := add(now.Add(-7*fullDay + time.Millisecond))
	changed := add(now.Add(-100 * fullDay))
	changeAt := now.Add(-8 * fullDay)
	setStatusAt(t, ctx, shop, changed.ID, StatusInspecting, changeAt)
	recent := add(now.Add(-100 * fullDay))
	setStatusAt(t, ctx, shop, recent.ID, StatusFinalTesting, now.Add(-time.Hour))
	delivered := add(now.Add(-100 * fullDay))
	setStatusAt(t, ctx, shop, delivered.ID, StatusDelivered, now.Add(-50*fullDay))
	cancelled := add(now.Add(-100 * fullDay))
	cancelAt := now.Add(-9 * fullDay)
	setStatusAt(t, ctx, shop, cancelled.ID, StatusCancelled, cancelAt)

	summary, err := LoadSummary(ctx, shop.db, now)
	if err != nil {
		t.Fatal(err)
	}
	if summary.Idle7 != 3 || summary.UnchangedMore || len(summary.Unchanged) != 5 {
		t.Fatalf("idle7 %d more %v rows %+v", summary.Idle7, summary.UnchangedMore, summary.Unchanged)
	}
	want := []struct {
		id     string
		days   int
		status string
		since  time.Time
	}{
		{cancelled.ID, 9, StatusCancelled, cancelAt},
		{changed.ID, 8, StatusInspecting, changeAt},
		{seven.ID, 7, StatusRegistered, now.Add(-7 * fullDay)},
		{almostSeven.ID, 6, StatusRegistered, now.Add(-7*fullDay + time.Millisecond)},
		{oneDay.ID, 1, StatusRegistered, now.Add(-fullDay)},
	}
	for i, item := range want {
		got := summary.Unchanged[i]
		if got.ID != item.id || got.Days != item.days || got.Status != item.status || !got.Since.Equal(item.since.UTC().Truncate(time.Millisecond)) || got.CustomerName != "Ada Lovelace" || got.Brand != "Idle" || got.Model != "Clock" {
			t.Fatalf("row %d %+v", i, got)
		}
	}
	for _, item := range summary.Unchanged {
		if item.ID == recent.ID || item.ID == delivered.ID {
			t.Fatalf("excluded device listed %+v", item)
		}
	}
}

func TestSummaryIdleLimit(t *testing.T) {
	ctx := context.Background()
	shop := openShop(t)
	now := summaryNow()
	var customer int64
	ids := make([]string, 0, 13)
	for days := 30; days >= 18; days-- {
		var id *int64
		if customer != 0 {
			id = &customer
		}
		device := registerAt(t, ctx, shop, id, "Idle", "Wait", now.Add(-time.Duration(days)*fullDay))
		if customer == 0 {
			customer = device.Customer.ID
		}
		ids = append(ids, device.ID)
	}

	summary, err := LoadSummary(ctx, shop.db, now)
	if err != nil {
		t.Fatal(err)
	}
	if !summary.UnchangedMore || summary.Idle7 != 13 || len(summary.Unchanged) != homeUnchanged {
		t.Fatalf("limit more %v idle %d len %d", summary.UnchangedMore, summary.Idle7, len(summary.Unchanged))
	}
	if summary.Unchanged[0].ID != ids[0] || summary.Unchanged[0].Days != 30 {
		t.Fatalf("longest %+v", summary.Unchanged[0])
	}
	if summary.Unchanged[11].ID != ids[11] || summary.Unchanged[11].Days != 19 {
		t.Fatalf("last shown %+v", summary.Unchanged[11])
	}
	for _, item := range summary.Unchanged {
		if item.ID == ids[12] {
			t.Fatal("shortest wait was shown")
		}
	}
}

func summaryNow() time.Time {
	return time.Date(2026, 11, 3, 9, 0, 0, 0, time.UTC)
}

func dayCount(days []DayCount, date string) int {
	for _, day := range days {
		if day.Date == date {
			return day.Count
		}
	}
	return -1
}

type shopFix struct {
	db    *sql.DB
	root  string
	actor User
}

func openShop(t *testing.T) shopFix {
	t.Helper()
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
	return shopFix{db: database, root: root, actor: tech}
}

func registerAt(t *testing.T, ctx context.Context, shop shopFix, customer *int64, brand, model string, at time.Time) Device {
	t.Helper()
	input := NewDevice{Brand: brand, Model: model, Note: "ok", CustomerID: customer}
	if customer == nil {
		input.CustomerName = "Ada Lovelace"
	}
	device, err := RegisterDevice(ctx, shop.db, shop.root, shop.actor, input)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := shop.db.ExecContext(ctx, `UPDATE devices SET created_at = ? WHERE id = ?`, timestamp(at), device.ID); err != nil {
		t.Fatal(err)
	}
	return device
}

func deliver(t *testing.T, ctx context.Context, shop shopFix, id string) {
	t.Helper()
	if _, err := ChangeDeviceStatus(ctx, shop.db, shop.root, shop.actor, id, StatusDelivered, "", nil); err != nil {
		t.Fatal(err)
	}
}

func setStatusAt(t *testing.T, ctx context.Context, shop shopFix, id, status string, at time.Time) {
	t.Helper()
	if _, err := ChangeDeviceStatus(ctx, shop.db, shop.root, shop.actor, id, status, "", nil); err != nil {
		t.Fatal(err)
	}
	if _, err := shop.db.ExecContext(ctx, `
		UPDATE device_status_changes
		SET created_at = ?
		WHERE device_id = ? AND id = (
			SELECT id FROM device_status_changes WHERE device_id = ? ORDER BY id DESC LIMIT 1
		)
	`, timestamp(at), id, id); err != nil {
		t.Fatal(err)
	}
}
