package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type ServiceComment struct {
	ID             int64
	AuthorID       int64
	AuthorUsername string
	Body           string
	CreatedAt      time.Time
	UpdatedAt      time.Time
	Edited         bool
}

func AddServiceComment(ctx context.Context, database *sql.DB, actor User, serviceID string, body string) (ServiceComment, error) {
	fields := &FieldError{}
	body = requiredText(body, maxNoteLen, "comment", fields)
	if err := fields.asError(); err != nil {
		return ServiceComment{}, err
	}
	if _, err := loadDevice(ctx, database, serviceID); err != nil {
		return ServiceComment{}, err
	}
	now := timestamp(time.Now())
	result, err := database.ExecContext(ctx, `
		INSERT INTO service_comments (service_id, author_id, author_username, body, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`, serviceID, actor.ID, actor.Username, body, now, now)
	if err != nil {
		return ServiceComment{}, fmt.Errorf("insert comment: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return ServiceComment{}, fmt.Errorf("comment id: %w", err)
	}
	return loadServiceComment(ctx, database, serviceID, id)
}

func UpdateServiceComment(ctx context.Context, database *sql.DB, actor User, serviceID string, commentID int64, body string) (ServiceComment, error) {
	fields := &FieldError{}
	body = requiredText(body, maxNoteLen, "comment", fields)
	if err := fields.asError(); err != nil {
		return ServiceComment{}, err
	}
	current, err := loadServiceComment(ctx, database, serviceID, commentID)
	if err != nil {
		return ServiceComment{}, err
	}
	if current.AuthorID != actor.ID {
		return ServiceComment{}, ErrForbidden
	}
	if current.Body == body {
		return current, nil
	}
	now := timestamp(time.Now())
	result, err := database.ExecContext(ctx, `
		UPDATE service_comments SET body = ?, updated_at = ?
		WHERE id = ? AND service_id = ? AND author_id = ?
	`, body, now, commentID, serviceID, actor.ID)
	if err != nil {
		return ServiceComment{}, fmt.Errorf("update comment: %w", err)
	}
	count, err := result.RowsAffected()
	if err != nil {
		return ServiceComment{}, fmt.Errorf("comment rows: %w", err)
	}
	if count != 1 {
		return ServiceComment{}, ErrNotFound
	}
	return loadServiceComment(ctx, database, serviceID, commentID)
}

func loadServiceComments(ctx context.Context, q queryer, serviceID string) ([]ServiceComment, error) {
	rows, err := q.QueryContext(ctx, `
		SELECT id, author_id, author_username, body, created_at, updated_at
		FROM service_comments
		WHERE service_id = ?
		ORDER BY created_at ASC, id ASC
	`, serviceID)
	if err != nil {
		return nil, fmt.Errorf("list comments: %w", err)
	}
	defer rows.Close()
	items := []ServiceComment{}
	for rows.Next() {
		item, err := scanServiceComment(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list comments: %w", err)
	}
	return items, nil
}

func loadServiceComment(ctx context.Context, q queryer, serviceID string, commentID int64) (ServiceComment, error) {
	row := q.QueryRowContext(ctx, `
		SELECT id, author_id, author_username, body, created_at, updated_at
		FROM service_comments
		WHERE id = ? AND service_id = ?
	`, commentID, serviceID)
	item, err := scanServiceComment(row)
	if err == sql.ErrNoRows {
		return ServiceComment{}, ErrNotFound
	}
	if err != nil {
		return ServiceComment{}, err
	}
	return item, nil
}

func scanServiceComment(row scanner) (ServiceComment, error) {
	var item ServiceComment
	var created string
	var updated string
	if err := row.Scan(&item.ID, &item.AuthorID, &item.AuthorUsername, &item.Body, &created, &updated); err != nil {
		return ServiceComment{}, err
	}
	parsedCreated, err := time.Parse(timestampLayout, created)
	if err != nil {
		return ServiceComment{}, fmt.Errorf("comment created: %w", err)
	}
	parsedUpdated, err := time.Parse(timestampLayout, updated)
	if err != nil {
		return ServiceComment{}, fmt.Errorf("comment updated: %w", err)
	}
	item.CreatedAt = parsedCreated
	item.UpdatedAt = parsedUpdated
	item.Edited = updated != created
	return item, nil
}
