package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

const (
	EventCustomerCreated      = "customer.created"
	EventCustomerUpdated      = "customer.updated"
	EventServiceRegistered    = "service.registered"
	EventServiceUpdated       = "service.updated"
	EventDeviceUpdated        = "device.updated"
	EventServiceStatusChanged = "service.status_changed"
	EventUserCreated          = "user.created"
	EventUserRoleChanged      = "user.role_changed"
	EventUserPasswordChanged  = "user.password_changed"
	EventUserActivated        = "user.activated"
	EventUserDeactivated      = "user.deactivated"
)

func knownEvent(event string) bool {
	switch event {
	case EventCustomerCreated, EventCustomerUpdated, EventServiceRegistered, EventServiceUpdated, EventDeviceUpdated, EventServiceStatusChanged,
		EventUserCreated, EventUserRoleChanged, EventUserPasswordChanged, EventUserActivated, EventUserDeactivated,
		EventStateAdded, EventStateUpdated, EventStatesReordered,
		EventSparePartAdded, EventSparePartUpdated, EventStockAdded, EventStockWrittenOff, EventSparePartUsed, EventSparePartReturned:
		return true
	default:
		return false
	}
}

func insertAudit(ctx context.Context, tx *sql.Tx, actor User, event string, subject string, detail any) error {
	body, err := json.Marshal(detail)
	if err != nil {
		return fmt.Errorf("audit detail: %w", err)
	}
	_, err = tx.ExecContext(ctx, `
		INSERT INTO audit_events (
			created_at, actor_id, actor_username, actor_username_fold, actor_role, event, subject_fold, detail
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, timestamp(time.Now()), actor.ID, actor.Username, fold(actor.Username), actor.Role, event, fold(subject), string(body))
	if err != nil {
		return fmt.Errorf("insert audit: %w", err)
	}
	return nil
}

type AuditQuery struct {
	ViewerRole string
	From       *time.Time
	To         *time.Time
	Person     string
	Event      string
	Subject    string
	Sort       string
	Limit      int
	AfterTime  string
	AfterID    int64
	HasAfter   bool
}

type AuditItem struct {
	ID            int64
	At            time.Time
	ActorUsername string
	ActorRole     string
	Event         string
	Detail        json.RawMessage
}

type AuditPage struct {
	Items   []AuditItem
	HasMore bool
}

func ListAudit(ctx context.Context, database *sql.DB, query AuditQuery) (AuditPage, error) {
	if query.ViewerRole == RoleTechnician || !validRole(query.ViewerRole) {
		return AuditPage{}, ErrForbidden
	}
	if query.Event != "" && !knownEvent(query.Event) {
		return AuditPage{}, ErrInvalidEvent
	}
	sort := query.Sort
	if sort == "" {
		sort = "desc"
	}
	if sort != "asc" && sort != "desc" {
		return AuditPage{}, ErrInvalidSort
	}
	limit, err := pageLimit(query.Limit)
	if err != nil {
		return AuditPage{}, err
	}

	clauses := []string{"1 = 1"}
	args := []any{}
	if query.ViewerRole == RoleOwner {
		clauses = append(clauses, "actor_role IN (?, ?)")
		args = append(args, RoleOwner, RoleTechnician)
	}
	if query.From != nil {
		clauses = append(clauses, "created_at >= ?")
		args = append(args, timestamp(*query.From))
	}
	if query.To != nil {
		clauses = append(clauses, "created_at < ?")
		args = append(args, timestamp(*query.To))
	}
	if strings.TrimSpace(query.Person) != "" {
		clauses = append(clauses, "actor_username_fold LIKE ? ESCAPE '\\'")
		args = append(args, likePattern(query.Person))
	}
	if query.Event != "" {
		clauses = append(clauses, "event = ?")
		args = append(args, query.Event)
	}
	if strings.TrimSpace(query.Subject) != "" {
		clauses = append(clauses, "subject_fold LIKE ? ESCAPE '\\'")
		args = append(args, likePattern(query.Subject))
	}
	if query.HasAfter {
		if sort == "asc" {
			clauses = append(clauses, "(created_at > ? OR (created_at = ? AND id > ?))")
		} else {
			clauses = append(clauses, "(created_at < ? OR (created_at = ? AND id < ?))")
		}
		args = append(args, query.AfterTime, query.AfterTime, query.AfterID)
	}

	order := "created_at DESC, id DESC"
	if sort == "asc" {
		order = "created_at ASC, id ASC"
	}
	args = append(args, limit+1)
	rows, err := database.QueryContext(ctx, `
		SELECT id, created_at, actor_username, actor_role, event, detail
		FROM audit_events
		WHERE `+strings.Join(clauses, " AND ")+`
		ORDER BY `+order+`
		LIMIT ?
	`, args...)
	if err != nil {
		return AuditPage{}, fmt.Errorf("list audit: %w", err)
	}
	defer rows.Close()

	var items []AuditItem
	for rows.Next() {
		var item AuditItem
		var at string
		var detail string
		if err := rows.Scan(&item.ID, &at, &item.ActorUsername, &item.ActorRole, &item.Event, &detail); err != nil {
			return AuditPage{}, fmt.Errorf("scan audit: %w", err)
		}
		parsed, err := time.Parse(timestampLayout, at)
		if err != nil {
			return AuditPage{}, fmt.Errorf("audit time: %w", err)
		}
		item.At = parsed
		item.Detail = json.RawMessage(detail)
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return AuditPage{}, fmt.Errorf("list audit: %w", err)
	}
	page := AuditPage{Items: items}
	if len(items) > limit {
		page.HasMore = true
		page.Items = items[:limit]
	}
	if page.Items == nil {
		page.Items = []AuditItem{}
	}
	return page, nil
}
