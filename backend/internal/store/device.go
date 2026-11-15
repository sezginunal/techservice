package store

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

type InShopError struct {
	ServiceID string
}

func (e *InShopError) Error() string {
	return "device_in_shop"
}

type Identity struct {
	ID        string
	Brand     string
	Model     string
	Serial    string
	CreatedAt time.Time
}

type EarlierService struct {
	ID           string
	CustomerName string
	Phone        string
	Status       string
	CreatedAt    time.Time
}

type DeviceService struct {
	ID           string
	CustomerName string
	Status       string
	CreatedAt    time.Time
}

type DeviceMatch struct {
	Matched       bool
	DeviceID      string
	OpenServiceID string
	Items         []EarlierService
	HasMore       bool
}

type VisitQuery struct {
	Limit     int
	AfterTime string
	AfterID   string
	HasAfter  bool
}

func MatchDevice(ctx context.Context, database *sql.DB, brand string, model string, serial string, query VisitQuery) (DeviceMatch, error) {
	fields := &FieldError{}
	brand = requiredText(brand, maxNameLen, "brand", fields)
	model = requiredText(model, maxNameLen, "model", fields)
	serial = requiredText(serial, maxSerialLen, "serial", fields)
	if err := fields.asError(); err != nil {
		return DeviceMatch{}, err
	}
	identity, err := findIdentity(ctx, database, brand, model, serial)
	if errorsIsNotFound(err) {
		return DeviceMatch{Items: []EarlierService{}}, nil
	}
	if err != nil {
		return DeviceMatch{}, err
	}
	limit, err := pageLimit(query.Limit)
	if err != nil {
		return DeviceMatch{}, err
	}
	items, hasMore, err := listEarlier(ctx, database, identity.ID, query, limit)
	if err != nil {
		return DeviceMatch{}, err
	}
	openID, err := openServiceID(ctx, database, identity.ID)
	if err != nil {
		return DeviceMatch{}, err
	}
	return DeviceMatch{
		Matched:       true,
		DeviceID:      identity.ID,
		OpenServiceID: openID,
		Items:         items,
		HasMore:       hasMore,
	}, nil
}

func GetIdentity(ctx context.Context, database *sql.DB, id string) (Identity, error) {
	var identity Identity
	var created string
	err := database.QueryRowContext(ctx, `
		SELECT id, brand, model, serial, created_at FROM devices WHERE id = ?
	`, id).Scan(&identity.ID, &identity.Brand, &identity.Model, &identity.Serial, &created)
	if err == sql.ErrNoRows {
		return Identity{}, ErrNotFound
	}
	if err != nil {
		return Identity{}, fmt.Errorf("read device: %w", err)
	}
	identity.CreatedAt, err = time.Parse(timestampLayout, created)
	if err != nil {
		return Identity{}, fmt.Errorf("device time: %w", err)
	}
	return identity, nil
}

func ListDeviceServices(ctx context.Context, database *sql.DB, deviceID string, query VisitQuery) ([]DeviceService, bool, error) {
	if _, err := GetIdentity(ctx, database, deviceID); err != nil {
		return nil, false, err
	}
	limit, err := pageLimit(query.Limit)
	if err != nil {
		return nil, false, err
	}
	clauses := []string{"services.device_id = ?"}
	args := []any{deviceID}
	if query.HasAfter {
		clauses = append(clauses, "(services.created_at < ? OR (services.created_at = ? AND services.id < ?))")
		args = append(args, query.AfterTime, query.AfterTime, query.AfterID)
	}
	args = append(args, limit+1)
	rows, err := database.QueryContext(ctx, `
		SELECT services.id, customers.name, services.status, services.created_at
		FROM services
		JOIN customers ON customers.id = services.customer_id
		WHERE `+strings.Join(clauses, " AND ")+`
		ORDER BY services.created_at DESC, services.id DESC
		LIMIT ?
	`, args...)
	if err != nil {
		return nil, false, fmt.Errorf("list device services: %w", err)
	}
	defer rows.Close()
	var items []DeviceService
	for rows.Next() {
		var item DeviceService
		var created string
		if err := rows.Scan(&item.ID, &item.CustomerName, &item.Status, &created); err != nil {
			return nil, false, fmt.Errorf("scan device service: %w", err)
		}
		item.CreatedAt, err = time.Parse(timestampLayout, created)
		if err != nil {
			return nil, false, fmt.Errorf("device service time: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, false, fmt.Errorf("list device services: %w", err)
	}
	hasMore := len(items) > limit
	if hasMore {
		items = items[:limit]
	}
	if items == nil {
		items = []DeviceService{}
	}
	return items, hasMore, nil
}

func findIdentity(ctx context.Context, q queryer, brand string, model string, serial string) (Identity, error) {
	var identity Identity
	var created string
	err := q.QueryRowContext(ctx, `
		SELECT id, brand, model, serial, created_at
		FROM devices
		WHERE brand_fold = ? AND model_fold = ? AND serial_fold = ?
	`, fold(brand), fold(model), fold(serial)).Scan(&identity.ID, &identity.Brand, &identity.Model, &identity.Serial, &created)
	if err == sql.ErrNoRows {
		return Identity{}, ErrNotFound
	}
	if err != nil {
		return Identity{}, fmt.Errorf("read device: %w", err)
	}
	identity.CreatedAt, err = time.Parse(timestampLayout, created)
	if err != nil {
		return Identity{}, fmt.Errorf("device time: %w", err)
	}
	return identity, nil
}

func ensureDevice(ctx context.Context, tx *sql.Tx, brand string, model string, serial string, now string) (string, error) {
	identity, err := findIdentity(ctx, tx, brand, model, serial)
	if err == nil {
		return identity.ID, nil
	}
	if !errorsIsNotFound(err) {
		return "", err
	}
	id, err := newID()
	if err != nil {
		return "", err
	}
	_, err = tx.ExecContext(ctx, `
		INSERT INTO devices (id, brand, model, serial, brand_fold, model_fold, serial_fold, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, id, brand, model, serial, fold(brand), fold(model), fold(serial), now)
	if err == nil {
		return id, nil
	}
	if !identityConflict(err) {
		return "", fmt.Errorf("insert device: %w", err)
	}
	identity, err = findIdentity(ctx, tx, brand, model, serial)
	if err != nil {
		return "", err
	}
	return identity.ID, nil
}

func openServiceID(ctx context.Context, q queryer, deviceID string) (string, error) {
	var id string
	err := q.QueryRowContext(ctx, `
		SELECT id FROM services
		WHERE device_id = ? AND status NOT IN (?, ?)
		ORDER BY created_at DESC, id DESC
		LIMIT 1
	`, deviceID, StatusDelivered, StatusCancelled).Scan(&id)
	if err == sql.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("open service: %w", err)
	}
	return id, nil
}

func listEarlier(ctx context.Context, database *sql.DB, deviceID string, query VisitQuery, limit int) ([]EarlierService, bool, error) {
	clauses := []string{"services.device_id = ?"}
	args := []any{deviceID}
	if query.HasAfter {
		clauses = append(clauses, "(services.created_at < ? OR (services.created_at = ? AND services.id < ?))")
		args = append(args, query.AfterTime, query.AfterTime, query.AfterID)
	}
	args = append(args, limit+1)
	rows, err := database.QueryContext(ctx, `
		SELECT services.id, customers.name, customers.phone, services.status, services.created_at
		FROM services
		JOIN customers ON customers.id = services.customer_id
		WHERE `+strings.Join(clauses, " AND ")+`
		ORDER BY services.created_at DESC, services.id DESC
		LIMIT ?
	`, args...)
	if err != nil {
		return nil, false, fmt.Errorf("list earlier services: %w", err)
	}
	defer rows.Close()
	var items []EarlierService
	for rows.Next() {
		var item EarlierService
		var created string
		if err := rows.Scan(&item.ID, &item.CustomerName, &item.Phone, &item.Status, &created); err != nil {
			return nil, false, fmt.Errorf("scan earlier service: %w", err)
		}
		item.CreatedAt, err = time.Parse(timestampLayout, created)
		if err != nil {
			return nil, false, fmt.Errorf("earlier service time: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, false, fmt.Errorf("list earlier services: %w", err)
	}
	hasMore := len(items) > limit
	if hasMore {
		items = items[:limit]
	}
	if items == nil {
		items = []EarlierService{}
	}
	return items, hasMore, nil
}

func errorsIsNotFound(err error) bool {
	return err != nil && err == ErrNotFound
}

func identityConflict(err error) bool {
	return err != nil && strings.Contains(err.Error(), "devices.brand_fold")
}

func openConflict(err error) bool {
	return err != nil && strings.Contains(err.Error(), "services_one_open")
}
