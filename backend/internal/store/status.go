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

func deviceStatusCheck() string {
	quoted := make([]string, 0, len(deviceStatusList()))
	for _, status := range deviceStatusList() {
		quoted = append(quoted, "'"+status+"'")
	}
	return "CHECK (status IN (" + strings.Join(quoted, ", ") + "))"
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
	active, err := activeStatus(ctx, database, next)
	if err != nil {
		return Device{}, err
	}
	if !active {
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
	stillActive, err := activeStatus(ctx, tx, next)
	if err != nil {
		return Device{}, err
	}
	if !stillActive {
		return Device{}, &FieldError{Invalid: []string{"status"}}
	}
	result, err := tx.ExecContext(ctx, `
		UPDATE services SET status = ? WHERE id = ? AND status = ?
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
		INSERT INTO service_status_changes (
			service_id, actor_username, from_status, to_status, note, photo_path, photo_size, photo_type, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, id, actor.Username, fresh.Status, next, note, photoPathValue, photoSize, photoType, timestamp(time.Now())); err != nil {
		return Device{}, fmt.Errorf("insert status change: %w", err)
	}
	if err := insertAudit(ctx, tx, actor, EventServiceStatusChanged, fresh.Brand+" "+fresh.Model+" "+fresh.Serial+" "+id+" "+fresh.Customer.Name, map[string]string{
		"brand":    fresh.Brand,
		"model":    fresh.Model,
		"serial":   fresh.Serial,
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
		FROM service_status_changes
		WHERE id = ? AND service_id = ? AND photo_path != ''
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
		FROM service_status_changes
		WHERE service_id = ?
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
