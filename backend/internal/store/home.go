package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"
	_ "time/tzdata"
)

const (
	homeDays      = 30
	homeWeek      = 7
	homeModels    = 8
	homeUnchanged = 12
	homeQueue     = 12
	homeLatest    = 5
	fullDay       = 24 * time.Hour
)

var istanbul = loadIstanbul()

func loadIstanbul() *time.Location {
	loc, err := time.LoadLocation("Europe/Istanbul")
	if err != nil {
		panic(err)
	}
	return loc
}

type DayCount struct {
	Date  string
	Count int
}

type StatusCount struct {
	Status string
	Count  int
}

type StateQueue struct {
	Status string
	More   bool
	Items  []IdleDevice
}

type ModelCount struct {
	Brand string
	Model string
	Count int
}

type IdleDevice struct {
	ID           string
	Brand        string
	Model        string
	Serial       string
	CustomerName string
	Status       string
	Days         int
	Since        time.Time
}

type RecentDevice struct {
	ID        string
	Brand     string
	Model     string
	Serial    string
	Status    string
	CreatedAt time.Time
}

type Summary struct {
	Registered7   int
	Registered30  int
	Days          []DayCount
	InShop        int
	Delivered     int
	Idle7         int
	ByStatus      []StatusCount
	Queues        []StateQueue
	Models        []ModelCount
	Other         int
	Unchanged     []IdleDevice
	UnchangedMore bool
	Latest        []RecentDevice
}

func LoadSummary(ctx context.Context, database *sql.DB, now time.Time) (Summary, error) {
	now = now.UTC().Truncate(time.Millisecond)
	summary := Summary{
		Days:      make([]DayCount, 0, homeDays),
		ByStatus:  []StatusCount{},
		Queues:    []StateQueue{},
		Models:    []ModelCount{},
		Unchanged: []IdleDevice{},
		Latest:    []RecentDevice{},
	}

	days, start, end := registrationDays(now)
	registered, err := registrationCounts(ctx, database, start, end)
	if err != nil {
		return Summary{}, err
	}
	var total int
	for i, day := range days {
		count := registered[day.Date]
		days[i].Count = count
		total += count
		if i >= homeDays-homeWeek {
			summary.Registered7 += count
		}
	}
	if total != countValues(registered) {
		return Summary{}, fmt.Errorf("home days: %d devices, %d in buckets", countValues(registered), total)
	}
	summary.Registered30 = total
	summary.Days = days

	if err := statusCounts(ctx, database, now, &summary); err != nil {
		return Summary{}, err
	}
	summary.Models, summary.Other, err = modelCounts(ctx, database, start, end)
	if err != nil {
		return Summary{}, err
	}
	if err := idleCounts(ctx, database, now, &summary); err != nil {
		return Summary{}, err
	}
	summary.Latest, err = latestDevices(ctx, database)
	if err != nil {
		return Summary{}, err
	}
	return summary, nil
}

func registrationDays(now time.Time) ([]DayCount, time.Time, time.Time) {
	local := now.In(istanbul)
	today := time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, istanbul)
	start := today.AddDate(0, 0, -(homeDays - 1))
	end := today.AddDate(0, 0, 1)
	days := make([]DayCount, homeDays)
	for i := range days {
		days[i].Date = start.AddDate(0, 0, i).Format("2006-01-02")
	}
	return days, start, end
}

func registrationCounts(ctx context.Context, database *sql.DB, start, end time.Time) (map[string]int, error) {
	rows, err := database.QueryContext(ctx, `
		SELECT created_at FROM services
		WHERE created_at >= ? AND created_at < ?
	`, timestamp(start), timestamp(end))
	if err != nil {
		return nil, fmt.Errorf("home registrations: %w", err)
	}
	defer rows.Close()

	counts := map[string]int{}
	for rows.Next() {
		var raw string
		if err := rows.Scan(&raw); err != nil {
			return nil, fmt.Errorf("home registration: %w", err)
		}
		parsed, err := time.Parse(timestampLayout, raw)
		if err != nil {
			return nil, fmt.Errorf("home registration time: %w", err)
		}
		counts[parsed.In(istanbul).Format("2006-01-02")]++
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("home registrations: %w", err)
	}
	return counts, nil
}

func countValues(counts map[string]int) int {
	total := 0
	for _, count := range counts {
		total += count
	}
	return total
}

func statusCounts(ctx context.Context, database *sql.DB, now time.Time, summary *Summary) error {
	rows, err := database.QueryContext(ctx, `SELECT status, COUNT(*) FROM services GROUP BY status`)
	if err != nil {
		return fmt.Errorf("home status: %w", err)
	}
	defer rows.Close()

	counts := map[string]int{}
	for rows.Next() {
		var status string
		var count int
		if err := rows.Scan(&status, &count); err != nil {
			return fmt.Errorf("home status: %w", err)
		}
		counts[status] = count
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("home status: %w", err)
	}
	rows.Close()

	states, err := database.QueryContext(ctx, `
		SELECT code, active FROM service_states ORDER BY position, code
	`)
	if err != nil {
		return fmt.Errorf("home states: %w", err)
	}
	defer states.Close()
	for states.Next() {
		var status string
		var active int
		if err := states.Scan(&status, &active); err != nil {
			return fmt.Errorf("home state: %w", err)
		}
		count := counts[status]
		if active == 0 && count == 0 {
			continue
		}
		summary.ByStatus = append(summary.ByStatus, StatusCount{Status: status, Count: count})
		if status == StatusDelivered {
			summary.Delivered = count
			continue
		}
		summary.InShop += count
	}
	if err := states.Err(); err != nil {
		return fmt.Errorf("home states: %w", err)
	}
	states.Close()
	summary.Queues, err = stateQueues(ctx, database, now)
	if err != nil {
		return err
	}
	return nil
}

func stateQueues(ctx context.Context, database *sql.DB, now time.Time) ([]StateQueue, error) {
	rows, err := database.QueryContext(ctx, `
		SELECT code FROM service_states WHERE show_on_home = 1 ORDER BY position, code
	`)
	if err != nil {
		return nil, fmt.Errorf("home queues: %w", err)
	}
	defer rows.Close()
	codes := []string{}
	for rows.Next() {
		var code string
		if err := rows.Scan(&code); err != nil {
			return nil, fmt.Errorf("home queue: %w", err)
		}
		codes = append(codes, code)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("home queues: %w", err)
	}
	queues := []StateQueue{}
	for _, code := range codes {
		queue, err := loadQueue(ctx, database, now, code)
		if err != nil {
			return nil, err
		}
		queues = append(queues, queue)
	}
	return queues, nil
}

func loadQueue(ctx context.Context, database *sql.DB, now time.Time, code string) (StateQueue, error) {
	rows, err := database.QueryContext(ctx, `
		SELECT services.id, devices.brand, devices.model, devices.serial, customers.name, services.status, `+sinceExpr+`
		FROM services
		JOIN devices ON devices.id = services.device_id
		JOIN customers ON customers.id = services.customer_id
		WHERE services.status = ?
		ORDER BY `+sinceExpr+` ASC, services.id ASC
		LIMIT ?
	`, code, homeQueue+1)
	if err != nil {
		return StateQueue{}, fmt.Errorf("home queue %s: %w", code, err)
	}
	defer rows.Close()
	queue := StateQueue{Status: code, Items: []IdleDevice{}}
	for rows.Next() {
		var item IdleDevice
		var raw string
		if err := rows.Scan(&item.ID, &item.Brand, &item.Model, &item.Serial, &item.CustomerName, &item.Status, &raw); err != nil {
			return StateQueue{}, fmt.Errorf("home queue %s: %w", code, err)
		}
		since, err := time.Parse(timestampLayout, raw)
		if err != nil {
			return StateQueue{}, fmt.Errorf("home queue time: %w", err)
		}
		item.Since = since
		item.Days = fullDays(now, since)
		queue.Items = append(queue.Items, item)
	}
	if err := rows.Err(); err != nil {
		return StateQueue{}, fmt.Errorf("home queue %s: %w", code, err)
	}
	if len(queue.Items) > homeQueue {
		queue.More = true
		queue.Items = queue.Items[:homeQueue]
	}
	return queue, nil
}

func modelCounts(ctx context.Context, database *sql.DB, start, end time.Time) ([]ModelCount, int, error) {
	rows, err := database.QueryContext(ctx, `
		SELECT devices.brand, devices.model, COUNT(*) AS n
		FROM services
		JOIN devices ON devices.id = services.device_id
		WHERE services.created_at >= ? AND services.created_at < ?
		GROUP BY devices.brand, devices.model
		ORDER BY n DESC, devices.brand COLLATE NOCASE, devices.model COLLATE NOCASE
	`, timestamp(start), timestamp(end))
	if err != nil {
		return nil, 0, fmt.Errorf("home models: %w", err)
	}
	defer rows.Close()

	var all []ModelCount
	for rows.Next() {
		var item ModelCount
		if err := rows.Scan(&item.Brand, &item.Model, &item.Count); err != nil {
			return nil, 0, fmt.Errorf("home model: %w", err)
		}
		all = append(all, item)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("home models: %w", err)
	}
	if len(all) <= homeModels {
		if all == nil {
			all = []ModelCount{}
		}
		return all, 0, nil
	}
	other := 0
	for _, item := range all[homeModels:] {
		other += item.Count
	}
	return all[:homeModels], other, nil
}

const sinceExpr = `COALESCE((SELECT MAX(created_at) FROM service_status_changes WHERE service_id = services.id), services.created_at)`

func idleCounts(ctx context.Context, database *sql.DB, now time.Time, summary *Summary) error {
	cutoff := timestamp(now.Add(-7 * fullDay))
	if err := database.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM services
		WHERE status NOT IN (?, ?)
		  AND `+sinceExpr+` <= ?
	`, StatusDelivered, StatusCancelled, cutoff).Scan(&summary.Idle7); err != nil {
		return fmt.Errorf("home idle: %w", err)
	}

	rows, err := database.QueryContext(ctx, `
		SELECT services.id, devices.brand, devices.model, devices.serial, customers.name, services.status, `+sinceExpr+`
		FROM services
		JOIN devices ON devices.id = services.device_id
		JOIN customers ON customers.id = services.customer_id
		WHERE services.status NOT IN (?, ?)
		  AND `+sinceExpr+` <= ?
		ORDER BY `+sinceExpr+` ASC, services.id ASC
		LIMIT ?
	`, StatusDelivered, StatusCancelled, timestamp(now.Add(-fullDay)), homeUnchanged+1)
	if err != nil {
		return fmt.Errorf("home unchanged: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var item IdleDevice
		var raw string
		if err := rows.Scan(&item.ID, &item.Brand, &item.Model, &item.Serial, &item.CustomerName, &item.Status, &raw); err != nil {
			return fmt.Errorf("home unchanged: %w", err)
		}
		since, err := time.Parse(timestampLayout, raw)
		if err != nil {
			return fmt.Errorf("home unchanged time: %w", err)
		}
		item.Since = since
		item.Days = fullDays(now, since)
		summary.Unchanged = append(summary.Unchanged, item)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("home unchanged: %w", err)
	}
	if len(summary.Unchanged) > homeUnchanged {
		summary.UnchangedMore = true
		summary.Unchanged = summary.Unchanged[:homeUnchanged]
	}
	return nil
}

func fullDays(now, since time.Time) int {
	elapsed := now.Sub(since)
	if elapsed < fullDay {
		return 0
	}
	return int(elapsed / fullDay)
}

func latestDevices(ctx context.Context, database *sql.DB) ([]RecentDevice, error) {
	rows, err := database.QueryContext(ctx, `
		SELECT services.id, devices.brand, devices.model, devices.serial, services.status, services.created_at
		FROM services
		JOIN devices ON devices.id = services.device_id
		ORDER BY services.created_at DESC, services.id DESC
		LIMIT ?
	`, homeLatest)
	if err != nil {
		return nil, fmt.Errorf("home latest: %w", err)
	}
	defer rows.Close()

	items := []RecentDevice{}
	for rows.Next() {
		var item RecentDevice
		var raw string
		if err := rows.Scan(&item.ID, &item.Brand, &item.Model, &item.Serial, &item.Status, &raw); err != nil {
			return nil, fmt.Errorf("home latest: %w", err)
		}
		parsed, err := time.Parse(timestampLayout, raw)
		if err != nil {
			return nil, fmt.Errorf("home latest time: %w", err)
		}
		item.CreatedAt = parsed
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("home latest: %w", err)
	}
	return items, nil
}
