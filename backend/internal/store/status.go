package store

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

const (
	StatusInspecting          = "inspecting"
	StatusRepairWarranty      = "repair_warranty"
	StatusRepairOutOfWarranty = "repair_out_of_warranty"
	StatusWaitingPartnerPart  = "waiting_partner_part"
	StatusWaitingOwnPart      = "waiting_own_part"
	StatusFinalTesting        = "final_testing"
	StatusPreparingDelivery   = "preparing_delivery"
	StatusInStorage           = "in_storage"
	StatusLeftStorage         = "left_storage"
	StatusDelivered           = "delivered"
	StatusCancelled           = "cancelled"
)

func deviceStatusList() []string {
	return []string{
		StatusRegistered,
		StatusInspecting,
		StatusRepairWarranty,
		StatusRepairOutOfWarranty,
		StatusWaitingPartnerPart,
		StatusWaitingOwnPart,
		StatusFinalTesting,
		StatusPreparingDelivery,
		StatusInStorage,
		StatusLeftStorage,
		StatusDelivered,
		StatusCancelled,
	}
}

func knownStatus(status string) bool {
	for _, item := range deviceStatusList() {
		if item == status {
			return true
		}
	}
	return false
}

func deviceStatusCheck() string {
	quoted := make([]string, 0, len(deviceStatusList()))
	for _, status := range deviceStatusList() {
		quoted = append(quoted, "'"+status+"'")
	}
	return "CHECK (status IN (" + strings.Join(quoted, ", ") + "))"
}

func widenDeviceStatus(ctx context.Context, database *sql.DB) error {
	var ddl string
	err := database.QueryRowContext(ctx, `SELECT sql FROM sqlite_master WHERE type = 'table' AND name = 'devices'`).Scan(&ddl)
	if err != nil {
		return fmt.Errorf("devices ddl: %w", err)
	}
	if strings.Contains(ddl, "'"+StatusCancelled+"'") {
		return nil
	}

	conn, err := database.Conn(ctx)
	if err != nil {
		return fmt.Errorf("devices conn: %w", err)
	}
	defer conn.Close()
	if _, err := conn.ExecContext(ctx, `PRAGMA foreign_keys = OFF`); err != nil {
		return fmt.Errorf("devices fk off: %w", err)
	}
	defer conn.ExecContext(context.Background(), `PRAGMA foreign_keys = ON`)

	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("devices migrate begin: %w", err)
	}
	defer tx.Rollback()
	script := fmt.Sprintf(`
		CREATE TABLE devices_next (
			id TEXT PRIMARY KEY,
			customer_id INTEGER NOT NULL REFERENCES customers(id),
			brand TEXT NOT NULL,
			model TEXT NOT NULL,
			serial TEXT NOT NULL DEFAULT '',
			note TEXT NOT NULL,
			status TEXT NOT NULL DEFAULT 'registered' %s,
			brand_fold TEXT NOT NULL,
			model_fold TEXT NOT NULL,
			serial_fold TEXT NOT NULL,
			created_at TEXT NOT NULL
		);
		INSERT INTO devices_next (
			id, customer_id, brand, model, serial, note, status,
			brand_fold, model_fold, serial_fold, created_at
		)
		SELECT id, customer_id, brand, model, serial, note, status,
			brand_fold, model_fold, serial_fold, created_at
		FROM devices;
		DROP TABLE devices;
		ALTER TABLE devices_next RENAME TO devices;
		CREATE INDEX devices_created ON devices (created_at, id);
		CREATE INDEX devices_customer ON devices (customer_id, created_at, id);
		CREATE INDEX devices_status ON devices (status, created_at, id);
	`, deviceStatusCheck())
	if _, err := tx.ExecContext(ctx, script); err != nil {
		return fmt.Errorf("devices migrate: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("devices migrate commit: %w", err)
	}
	return nil
}

type StatusChange struct {
	ID            int64
	ActorUsername string
	From          string
	To            string
	Note          string
	HasPhoto      bool
	CreatedAt     time.Time
}

func ChangeDeviceStatus(ctx context.Context, database *sql.DB, root string, actor User, id string, next string, note string, file *Upload) (Device, error) {
	fields := &FieldError{}
	if !knownStatus(next) {
		fields.addInvalid("status")
	}
	note = optionalText(note, maxNoteLen, "statusNote", fields)
	if err := fields.asError(); err != nil {
		return Device{}, err
	}

	current, err := loadDevice(ctx, database, id)
	if err != nil {
		return Device{}, err
	}
	if current.Status == next {
		return Device{}, &FieldError{Invalid: []string{"status"}}
	}

	var stored *storedPhoto
	if file != nil {
		body, err := file.Open()
		if err != nil {
			return Device{}, fmt.Errorf("open status photo: %w", err)
		}
		photo, err := writePhoto(root, id, body)
		body.Close()
		if err != nil {
			return Device{}, err
		}
		stored = &photo
	}
	committed := false
	if stored != nil {
		defer func() {
			if !committed {
				removePaths([]string{photoPath(root, stored.rel)})
			}
		}()
	}

	tx, err := database.BeginTx(ctx, nil)
	if err != nil {
		return Device{}, fmt.Errorf("begin status: %w", err)
	}
	defer tx.Rollback()

	fresh, err := loadDevice(ctx, tx, id)
	if err != nil {
		return Device{}, err
	}
	if fresh.Status == next {
		return Device{}, &FieldError{Invalid: []string{"status"}}
	}
	result, err := tx.ExecContext(ctx, `
		UPDATE devices SET status = ? WHERE id = ? AND status = ?
	`, next, id, fresh.Status)
	if err != nil {
		return Device{}, fmt.Errorf("update status: %w", err)
	}
	count, err := result.RowsAffected()
	if err != nil {
		return Device{}, fmt.Errorf("status rows: %w", err)
	}
	if count != 1 {
		return Device{}, &FieldError{Invalid: []string{"status"}}
	}

	photoPathValue := ""
	photoSize := int64(0)
	photoType := ""
	if stored != nil {
		photoPathValue = stored.rel
		photoSize = stored.size
		photoType = stored.media
	}
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO device_status_changes (
			device_id, actor_username, from_status, to_status, note, photo_path, photo_size, photo_type, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, id, actor.Username, fresh.Status, next, note, photoPathValue, photoSize, photoType, timestamp(time.Now())); err != nil {
		return Device{}, fmt.Errorf("insert status change: %w", err)
	}
	if err := insertAudit(ctx, tx, actor, EventDeviceStatusChanged, fresh.Brand+" "+fresh.Model+" "+id+" "+fresh.Customer.Name, map[string]string{
		"brand":    fresh.Brand,
		"model":    fresh.Model,
		"uuid":     id,
		"customer": fresh.Customer.Name,
		"from":     fresh.Status,
		"to":       next,
	}); err != nil {
		return Device{}, err
	}
	if err := tx.Commit(); err != nil {
		return Device{}, fmt.Errorf("commit status: %w", err)
	}
	committed = true
	return loadDevice(ctx, database, id)
}

func OpenStatusPhoto(ctx context.Context, database *sql.DB, root string, deviceID string, changeID int64) (string, string, error) {
	var rel string
	var media string
	err := database.QueryRowContext(ctx, `
		SELECT photo_path, photo_type
		FROM device_status_changes
		WHERE id = ? AND device_id = ? AND photo_path != ''
	`, changeID, deviceID).Scan(&rel, &media)
	if err == sql.ErrNoRows {
		return "", "", ErrNotFound
	}
	if err != nil {
		return "", "", fmt.Errorf("read status photo: %w", err)
	}
	full, err := photoFull(root, rel)
	if err != nil {
		return "", "", err
	}
	return full, media, nil
}

func loadStatusHistory(ctx context.Context, q queryer, deviceID string) ([]StatusChange, error) {
	rows, err := q.QueryContext(ctx, `
		SELECT id, actor_username, from_status, to_status, note, photo_path, created_at
		FROM device_status_changes
		WHERE device_id = ?
		ORDER BY created_at DESC, id DESC
	`, deviceID)
	if err != nil {
		return nil, fmt.Errorf("list status changes: %w", err)
	}
	defer rows.Close()
	items := []StatusChange{}
	for rows.Next() {
		var item StatusChange
		var photo string
		var created string
		if err := rows.Scan(&item.ID, &item.ActorUsername, &item.From, &item.To, &item.Note, &photo, &created); err != nil {
			return nil, fmt.Errorf("scan status change: %w", err)
		}
		item.HasPhoto = photo != ""
		parsed, err := time.Parse(timestampLayout, created)
		if err != nil {
			return nil, fmt.Errorf("status time: %w", err)
		}
		item.CreatedAt = parsed
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list status changes: %w", err)
	}
	return items, nil
}

func photoPath(root string, rel string) string {
	full, err := photoFull(root, rel)
	if err != nil {
		return ""
	}
	return full
}
