package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

const (
	maxStock = 99999

	moveReceipt  = "receipt"
	moveWriteOff = "write_off"
	moveUse      = "use"
	moveReturn   = "return"

	EventSparePartAdded    = "spare_part.added"
	EventSparePartUpdated  = "spare_part.updated"
	EventStockAdded        = "stock.added"
	EventStockWrittenOff   = "stock.written_off"
	EventSparePartUsed     = "spare_part.used"
	EventSparePartReturned = "spare_part.returned"

	StockShort  = "stock_short"
	StockCap    = "stock_cap"
	ReturnShort = "return_short"

	suggestLimit = 50
)

type StockError struct {
	Kind       string
	OnHand     int
	Returnable int
}

func (e *StockError) Error() string {
	return e.Kind
}

type SparePart struct {
	ID     string
	Brand  string
	Model  string
	OnHand int
}

type SpareCreated struct {
	SparePart
	Existed bool
	Added   bool
}

type SpareMove struct {
	ID            int64
	At            time.Time
	ActorUsername string
	Quantity      int
	Kind          string
	Note          string
	Status        string
	ServiceID     string
	PartID        string
	Brand         string
	Model         string
	DeviceBrand   string
	DeviceModel   string
	Serial        string
}

type SpareReturnable struct {
	ID         string
	Brand      string
	Model      string
	Returnable int
}

type SpareListQuery struct {
	Text       string
	Limit      int
	AfterBrand string
	AfterModel string
	AfterID    string
	HasAfter   bool
}

type SpareMoveQuery struct {
	Limit     int
	AfterTime string
	AfterID   int64
	HasAfter  bool
}

type SparePage struct {
	Items   []SparePart
	HasMore bool
}

type SpareMovePage struct {
	Items   []SpareMove
	HasMore bool
}

type SpareReturnPage struct {
	Items   []SpareReturnable
	HasMore bool
}

type spareDB interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
	ExecContext(context.Context, string, ...any) (sql.Result, error)
}

type serviceDevice struct {
	Status string
	Brand  string
	Model  string
	Serial string
}

func CreateSparePart(ctx context.Context, database *sql.DB, actor User, brand string, model string, quantity *float64) (SpareCreated, error) {
	fields := &FieldError{}
	brand, model = spareNames(brand, model, fields)
	qty := quantityIn(quantity, 0, maxStock, fields)
	if err := fields.asError(); err != nil {
		return SpareCreated{}, err
	}

	tx, err := database.BeginTx(ctx, nil)
	if err != nil {
		return SpareCreated{}, fmt.Errorf("begin spare part: %w", err)
	}
	defer tx.Rollback()

	created, err := insertSpare(ctx, tx, actor, brand, model, qty)
	if err != nil {
		return SpareCreated{}, err
	}
	if err := tx.Commit(); err != nil {
		return SpareCreated{}, fmt.Errorf("commit spare part: %w", err)
	}
	return created, nil
}

func UpdateSparePart(ctx context.Context, database *sql.DB, actor User, id string, brand string, model string) (SparePart, error) {
	fields := &FieldError{}
	brand, model = spareNames(brand, model, fields)
	if err := fields.asError(); err != nil {
		return SparePart{}, err
	}

	tx, err := database.BeginTx(ctx, nil)
	if err != nil {
		return SparePart{}, fmt.Errorf("begin spare part: %w", err)
	}
	defer tx.Rollback()

	current, err := loadSpare(ctx, tx, id)
	if err != nil {
		return SparePart{}, err
	}
	if brand == current.Brand && model == current.Model {
		return current, nil
	}
	var other string
	err = tx.QueryRowContext(ctx, `
		SELECT id FROM spare_parts WHERE brand_fold = ? AND model_fold = ? AND id <> ?
	`, fold(brand), fold(model), id).Scan(&other)
	if err == nil {
		return SparePart{}, &FieldError{Taken: []string{"spare"}}
	}
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return SparePart{}, fmt.Errorf("match spare part: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `
		UPDATE spare_parts SET brand = ?, model = ?, brand_fold = ?, model_fold = ? WHERE id = ?
	`, brand, model, fold(brand), fold(model), id); err != nil {
		if spareIdentityConflict(err) {
			return SparePart{}, &FieldError{Taken: []string{"spare"}}
		}
		return SparePart{}, fmt.Errorf("update spare part: %w", err)
	}
	if err := insertAudit(ctx, tx, actor, EventSparePartUpdated, current.Brand+" "+current.Model+" "+brand+" "+model, map[string]string{
		"previousBrand": current.Brand,
		"previousModel": current.Model,
		"brand":         brand,
		"model":         model,
	}); err != nil {
		return SparePart{}, err
	}
	if err := tx.Commit(); err != nil {
		return SparePart{}, fmt.Errorf("commit spare part: %w", err)
	}
	return loadSpare(ctx, database, id)
}

func AddSpareStock(ctx context.Context, database *sql.DB, actor User, id string, quantity *float64, note string) (SparePart, error) {
	fields := &FieldError{}
	qty := quantityIn(quantity, 1, maxStock, fields)
	note = optionalText(note, maxNoteLen, "spareNote", fields)
	if err := fields.asError(); err != nil {
		return SparePart{}, err
	}
	return changeSpareStock(ctx, database, actor, id, qty, note, moveReceipt)
}

func WriteOffSpareStock(ctx context.Context, database *sql.DB, actor User, id string, quantity *float64, note string) (SparePart, error) {
	fields := &FieldError{}
	qty := quantityIn(quantity, 1, maxStock, fields)
	note = optionalText(note, maxNoteLen, "spareNote", fields)
	if err := fields.asError(); err != nil {
		return SparePart{}, err
	}
	return changeSpareStock(ctx, database, actor, id, qty, note, moveWriteOff)
}

func UseSparePart(ctx context.Context, database *sql.DB, actor User, serviceID string, partID string, quantity *float64) (int, error) {
	fields := &FieldError{}
	qty := quantityIn(quantity, 1, maxStock, fields)
	if err := fields.asError(); err != nil {
		return 0, err
	}
	return moveSpareOnService(ctx, database, actor, serviceID, partID, qty, "", moveUse)
}

func ReturnSparePart(ctx context.Context, database *sql.DB, actor User, serviceID string, partID string, quantity *float64, note string) (int, error) {
	fields := &FieldError{}
	qty := quantityIn(quantity, 1, maxStock, fields)
	note = optionalText(note, maxNoteLen, "spareNote", fields)
	if err := fields.asError(); err != nil {
		return 0, err
	}
	return moveSpareOnService(ctx, database, actor, serviceID, partID, qty, note, moveReturn)
}

func GetSparePart(ctx context.Context, database *sql.DB, id string) (SparePart, error) {
	return loadSpare(ctx, database, id)
}

func ListSpareParts(ctx context.Context, database *sql.DB, query SpareListQuery) (SparePage, error) {
	limit, err := pageLimit(query.Limit)
	if err != nil {
		return SparePage{}, err
	}
	clauses := []string{}
	args := []any{}
	if strings.TrimSpace(query.Text) != "" {
		pattern := likePattern(query.Text)
		clauses = append(clauses, "(brand_fold LIKE ? ESCAPE '\\' OR model_fold LIKE ? ESCAPE '\\')")
		args = append(args, pattern, pattern)
	}
	if query.HasAfter {
		clauses = append(clauses, spareOrderAfter("brand", "model", "id"))
		args = append(args, query.AfterBrand, query.AfterBrand, query.AfterModel, query.AfterBrand, query.AfterModel, query.AfterID)
	}
	where := ""
	if len(clauses) > 0 {
		where = "WHERE " + strings.Join(clauses, " AND ")
	}
	args = append(args, limit+1)
	rows, err := database.QueryContext(ctx, `
		SELECT id, brand, model, on_hand
		FROM spare_parts
		`+where+`
		ORDER BY brand COLLATE NOCASE, model COLLATE NOCASE, id
		LIMIT ?
	`, args...)
	if err != nil {
		return SparePage{}, fmt.Errorf("list spare parts: %w", err)
	}
	defer rows.Close()

	var items []SparePart
	for rows.Next() {
		var item SparePart
		if err := rows.Scan(&item.ID, &item.Brand, &item.Model, &item.OnHand); err != nil {
			return SparePage{}, fmt.Errorf("scan spare part: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return SparePage{}, fmt.Errorf("list spare parts: %w", err)
	}
	items, more := pageCut(items, limit)
	return SparePage{Items: items, HasMore: more}, nil
}

func SuggestSpareBrands(ctx context.Context, database *sql.DB, text string) ([]string, error) {
	rows, err := database.QueryContext(ctx, `
		SELECT MIN(brand)
		FROM spare_parts
		WHERE brand_fold LIKE ? ESCAPE '\'
		GROUP BY brand_fold
		ORDER BY MIN(brand) COLLATE NOCASE
		LIMIT ?
	`, likePattern(text), suggestLimit)
	if err != nil {
		return nil, fmt.Errorf("suggest spare brands: %w", err)
	}
	defer rows.Close()
	return scanNames(rows)
}

func SuggestSpareModels(ctx context.Context, database *sql.DB, brand string, text string) ([]string, error) {
	if strings.TrimSpace(brand) == "" {
		return []string{}, nil
	}
	rows, err := database.QueryContext(ctx, `
		SELECT model
		FROM spare_parts
		WHERE brand_fold = ? AND model_fold LIKE ? ESCAPE '\'
		ORDER BY model COLLATE NOCASE
		LIMIT ?
	`, fold(brand), likePattern(text), suggestLimit)
	if err != nil {
		return nil, fmt.Errorf("suggest spare models: %w", err)
	}
	defer rows.Close()
	return scanNames(rows)
}

func ListSpareUses(ctx context.Context, database *sql.DB, partID string, query SpareMoveQuery) (SpareMovePage, error) {
	if _, err := loadSpare(ctx, database, partID); err != nil {
		return SpareMovePage{}, err
	}
	return listMoves(ctx, database, query, `
		SELECT moves.id, moves.created_at, moves.actor_username, moves.quantity, moves.kind, moves.note,
			moves.state_code, moves.service_id, moves.part_id, '', '', devices.brand, devices.model, devices.serial
		FROM spare_part_moves moves
		JOIN services ON services.id = moves.service_id
		JOIN devices ON devices.id = services.device_id
		WHERE moves.part_id = ? AND moves.kind IN (?, ?)
	`, []any{partID, moveUse, moveReturn}, "moves")
}

func ListSpareStock(ctx context.Context, database *sql.DB, partID string, query SpareMoveQuery) (SpareMovePage, error) {
	if _, err := loadSpare(ctx, database, partID); err != nil {
		return SpareMovePage{}, err
	}
	return listMoves(ctx, database, query, `
		SELECT id, created_at, actor_username, quantity, kind, note,
			'', '', '', '', '', '', '', ''
		FROM spare_part_moves
		WHERE part_id = ? AND kind IN (?, ?)
	`, []any{partID, moveReceipt, moveWriteOff}, "")
}

func ListServiceSpares(ctx context.Context, database *sql.DB, serviceID string, query SpareMoveQuery) (SpareMovePage, error) {
	if err := requireService(ctx, database, serviceID); err != nil {
		return SpareMovePage{}, err
	}
	return listMoves(ctx, database, query, `
		SELECT moves.id, moves.created_at, moves.actor_username, moves.quantity, moves.kind, moves.note,
			moves.state_code, moves.service_id, spare_parts.id, spare_parts.brand, spare_parts.model, '', '', ''
		FROM spare_part_moves moves
		JOIN spare_parts ON spare_parts.id = moves.part_id
		WHERE moves.service_id = ? AND moves.kind IN (?, ?)
	`, []any{serviceID, moveUse, moveReturn}, "moves")
}

func ListReturnableSpares(ctx context.Context, database *sql.DB, serviceID string, query SpareListQuery) (SpareReturnPage, error) {
	if err := requireService(ctx, database, serviceID); err != nil {
		return SpareReturnPage{}, err
	}
	limit, err := pageLimit(query.Limit)
	if err != nil {
		return SpareReturnPage{}, err
	}
	having := []string{"returnable >= 1"}
	args := []any{serviceID, moveUse, moveReturn}
	if query.HasAfter {
		having = append(having, spareOrderAfter("p.brand", "p.model", "p.id"))
		args = append(args, query.AfterBrand, query.AfterBrand, query.AfterModel, query.AfterBrand, query.AfterModel, query.AfterID)
	}
	args = append(args, limit+1)
	rows, err := database.QueryContext(ctx, `
		SELECT p.id, p.brand, p.model,
			SUM(CASE moves.kind WHEN ? THEN moves.quantity ELSE -moves.quantity END) AS returnable
		FROM spare_part_moves moves
		JOIN spare_parts p ON p.id = moves.part_id
		WHERE moves.service_id = ? AND moves.kind IN (?, ?)
		GROUP BY p.id
		HAVING `+strings.Join(having, " AND ")+`
		ORDER BY p.brand COLLATE NOCASE, p.model COLLATE NOCASE, p.id
		LIMIT ?
	`, append([]any{moveUse}, args...)...)
	if err != nil {
		return SpareReturnPage{}, fmt.Errorf("list returnable spares: %w", err)
	}
	defer rows.Close()

	var items []SpareReturnable
	for rows.Next() {
		var item SpareReturnable
		if err := rows.Scan(&item.ID, &item.Brand, &item.Model, &item.Returnable); err != nil {
			return SpareReturnPage{}, fmt.Errorf("scan returnable spare: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return SpareReturnPage{}, fmt.Errorf("list returnable spares: %w", err)
	}
	items, more := pageCut(items, limit)
	return SpareReturnPage{Items: items, HasMore: more}, nil
}

func insertSpare(ctx context.Context, tx *sql.Tx, actor User, brand string, model string, qty int) (SpareCreated, error) {
	existing, err := findSpare(ctx, tx, brand, model)
	if err == nil {
		return finishExisting(ctx, tx, actor, existing, qty)
	}
	if !errors.Is(err, ErrNotFound) {
		return SpareCreated{}, err
	}
	id, err := newID()
	if err != nil {
		return SpareCreated{}, err
	}
	_, err = tx.ExecContext(ctx, `
		INSERT INTO spare_parts (id, brand, model, brand_fold, model_fold, on_hand, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, id, brand, model, fold(brand), fold(model), qty, timestamp(time.Now()))
	if spareIdentityConflict(err) {
		existing, err = findSpare(ctx, tx, brand, model)
		if err != nil {
			return SpareCreated{}, err
		}
		return finishExisting(ctx, tx, actor, existing, qty)
	}
	if err != nil {
		return SpareCreated{}, fmt.Errorf("insert spare part: %w", err)
	}
	if qty > 0 {
		if err := insertMove(ctx, tx, id, moveReceipt, qty, "", "", "", actor.Username); err != nil {
			return SpareCreated{}, err
		}
	}
	if err := insertAudit(ctx, tx, actor, EventSparePartAdded, brand+" "+model+" "+strconv.Itoa(qty), map[string]any{
		"brand":    brand,
		"model":    model,
		"quantity": qty,
	}); err != nil {
		return SpareCreated{}, err
	}
	part, err := loadSpare(ctx, tx, id)
	if err != nil {
		return SpareCreated{}, err
	}
	return SpareCreated{SparePart: part}, nil
}

func finishExisting(ctx context.Context, tx *sql.Tx, actor User, existing SparePart, qty int) (SpareCreated, error) {
	if qty == 0 {
		return SpareCreated{SparePart: existing, Existed: true}, nil
	}
	if err := raiseStock(ctx, tx, existing.ID, qty); err != nil {
		return SpareCreated{}, err
	}
	if err := insertMove(ctx, tx, existing.ID, moveReceipt, qty, "", "", "", actor.Username); err != nil {
		return SpareCreated{}, err
	}
	part, err := loadSpare(ctx, tx, existing.ID)
	if err != nil {
		return SpareCreated{}, err
	}
	if err := insertAudit(ctx, tx, actor, EventStockAdded, part.Brand+" "+part.Model+" "+strconv.Itoa(qty), map[string]any{
		"brand":    part.Brand,
		"model":    part.Model,
		"quantity": qty,
	}); err != nil {
		return SpareCreated{}, err
	}
	return SpareCreated{SparePart: part, Existed: true, Added: true}, nil
}

func changeSpareStock(ctx context.Context, database *sql.DB, actor User, id string, qty int, note string, kind string) (SparePart, error) {
	tx, err := database.BeginTx(ctx, nil)
	if err != nil {
		return SparePart{}, fmt.Errorf("begin spare stock: %w", err)
	}
	defer tx.Rollback()

	if _, err := loadSpare(ctx, tx, id); err != nil {
		return SparePart{}, err
	}
	if kind == moveReceipt {
		if err := raiseStock(ctx, tx, id, qty); err != nil {
			return SparePart{}, err
		}
	} else if err := lowerStock(ctx, tx, id, qty); err != nil {
		return SparePart{}, err
	}
	if err := insertMove(ctx, tx, id, kind, qty, note, "", "", actor.Username); err != nil {
		return SparePart{}, err
	}
	part, err := loadSpare(ctx, tx, id)
	if err != nil {
		return SparePart{}, err
	}
	event := EventStockAdded
	if kind == moveWriteOff {
		event = EventStockWrittenOff
	}
	if err := insertAudit(ctx, tx, actor, event, part.Brand+" "+part.Model+" "+strconv.Itoa(qty), map[string]any{
		"brand":    part.Brand,
		"model":    part.Model,
		"quantity": qty,
	}); err != nil {
		return SparePart{}, err
	}
	if err := tx.Commit(); err != nil {
		return SparePart{}, fmt.Errorf("commit spare stock: %w", err)
	}
	return part, nil
}

func moveSpareOnService(ctx context.Context, database *sql.DB, actor User, serviceID string, partID string, qty int, note string, kind string) (int, error) {
	tx, err := database.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("begin spare use: %w", err)
	}
	defer tx.Rollback()

	part, err := loadSpare(ctx, tx, partID)
	if err != nil {
		return 0, err
	}
	service, err := loadServiceDevice(ctx, tx, serviceID)
	if err != nil {
		return 0, err
	}
	if kind == moveReturn {
		returnable, err := returnableQty(ctx, tx, partID, serviceID)
		if err != nil {
			return 0, err
		}
		if qty > returnable {
			return 0, &StockError{Kind: ReturnShort, Returnable: returnable}
		}
		if err := raiseStock(ctx, tx, partID, qty); err != nil {
			return 0, err
		}
	} else if err := lowerStock(ctx, tx, partID, qty); err != nil {
		return 0, err
	}
	if err := insertMove(ctx, tx, partID, kind, qty, note, serviceID, service.Status, actor.Username); err != nil {
		return 0, err
	}
	part, err = loadSpare(ctx, tx, partID)
	if err != nil {
		return 0, err
	}
	event := EventSparePartUsed
	if kind == moveReturn {
		event = EventSparePartReturned
	}
	subject := strings.Join([]string{
		part.Brand, part.Model, strconv.Itoa(qty), serviceID, service.Brand, service.Model, service.Serial, service.Status,
	}, " ")
	if err := insertAudit(ctx, tx, actor, event, subject, map[string]any{
		"brand":       part.Brand,
		"model":       part.Model,
		"quantity":    qty,
		"uuid":        serviceID,
		"deviceBrand": service.Brand,
		"deviceModel": service.Model,
		"serial":      service.Serial,
		"status":      service.Status,
	}); err != nil {
		return 0, err
	}
	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("commit spare use: %w", err)
	}
	return part.OnHand, nil
}

func raiseStock(ctx context.Context, tx *sql.Tx, id string, quantity int) error {
	result, err := tx.ExecContext(ctx, `
		UPDATE spare_parts SET on_hand = on_hand + ? WHERE id = ? AND on_hand <= ?
	`, quantity, id, maxStock-quantity)
	if err != nil {
		return fmt.Errorf("raise stock: %w", err)
	}
	count, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("raise stock: %w", err)
	}
	if count == 1 {
		return nil
	}
	var onHand int
	err = tx.QueryRowContext(ctx, `SELECT on_hand FROM spare_parts WHERE id = ?`, id).Scan(&onHand)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	if err != nil {
		return fmt.Errorf("read stock: %w", err)
	}
	return &StockError{Kind: StockCap, OnHand: onHand}
}

func lowerStock(ctx context.Context, tx *sql.Tx, id string, quantity int) error {
	result, err := tx.ExecContext(ctx, `
		UPDATE spare_parts SET on_hand = on_hand - ? WHERE id = ? AND on_hand >= ?
	`, quantity, id, quantity)
	if err != nil {
		return fmt.Errorf("lower stock: %w", err)
	}
	count, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("lower stock: %w", err)
	}
	if count == 1 {
		return nil
	}
	var onHand int
	err = tx.QueryRowContext(ctx, `SELECT on_hand FROM spare_parts WHERE id = ?`, id).Scan(&onHand)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	if err != nil {
		return fmt.Errorf("read stock: %w", err)
	}
	return &StockError{Kind: StockShort, OnHand: onHand}
}

func returnableQty(ctx context.Context, tx *sql.Tx, partID string, serviceID string) (int, error) {
	var qty int
	err := tx.QueryRowContext(ctx, `
		SELECT COALESCE(SUM(CASE kind WHEN ? THEN quantity ELSE -quantity END), 0)
		FROM spare_part_moves
		WHERE part_id = ? AND service_id = ? AND kind IN (?, ?)
	`, moveUse, partID, serviceID, moveUse, moveReturn).Scan(&qty)
	if err != nil {
		return 0, fmt.Errorf("returnable stock: %w", err)
	}
	return qty, nil
}

func insertMove(ctx context.Context, tx *sql.Tx, partID string, kind string, quantity int, note string, serviceID string, state string, actor string) error {
	var service any
	var stateCode any
	if serviceID != "" {
		service = serviceID
		stateCode = state
	}
	_, err := tx.ExecContext(ctx, `
		INSERT INTO spare_part_moves (
			part_id, kind, quantity, note, service_id, state_code, actor_username, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, partID, kind, quantity, note, service, stateCode, actor, timestamp(time.Now()))
	if err != nil {
		return fmt.Errorf("insert spare move: %w", err)
	}
	return nil
}

func listMoves(ctx context.Context, database *sql.DB, query SpareMoveQuery, base string, args []any, alias string) (SpareMovePage, error) {
	limit, err := pageLimit(query.Limit)
	if err != nil {
		return SpareMovePage{}, err
	}
	prefix := ""
	if alias != "" {
		prefix = alias + "."
	}
	if query.HasAfter {
		base += fmt.Sprintf(" AND (%screated_at < ? OR (%screated_at = ? AND %sid < ?))", prefix, prefix, prefix)
		args = append(args, query.AfterTime, query.AfterTime, query.AfterID)
	}
	args = append(args, limit+1)
	rows, err := database.QueryContext(ctx, base+fmt.Sprintf(" ORDER BY %screated_at DESC, %sid DESC LIMIT ?", prefix, prefix), args...)
	if err != nil {
		return SpareMovePage{}, fmt.Errorf("list spare moves: %w", err)
	}
	defer rows.Close()

	var items []SpareMove
	for rows.Next() {
		var item SpareMove
		var at string
		var state sql.NullString
		var service sql.NullString
		if err := rows.Scan(
			&item.ID, &at, &item.ActorUsername, &item.Quantity, &item.Kind, &item.Note,
			&state, &service, &item.PartID, &item.Brand, &item.Model, &item.DeviceBrand, &item.DeviceModel, &item.Serial,
		); err != nil {
			return SpareMovePage{}, fmt.Errorf("scan spare move: %w", err)
		}
		parsed, err := time.Parse(timestampLayout, at)
		if err != nil {
			return SpareMovePage{}, fmt.Errorf("spare move time: %w", err)
		}
		item.At = parsed
		item.Status = state.String
		item.ServiceID = service.String
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return SpareMovePage{}, fmt.Errorf("list spare moves: %w", err)
	}
	items, more := pageCut(items, limit)
	return SpareMovePage{Items: items, HasMore: more}, nil
}

func spareOrderAfter(brand string, model string, id string) string {
	return fmt.Sprintf(
		"(%s COLLATE NOCASE > ? OR (%s COLLATE NOCASE = ? AND %s COLLATE NOCASE > ?) OR (%s COLLATE NOCASE = ? AND %s COLLATE NOCASE = ? AND %s > ?))",
		brand, brand, model, brand, model, id,
	)
}

func loadSpare(ctx context.Context, q spareDB, id string) (SparePart, error) {
	var part SparePart
	err := q.QueryRowContext(ctx, `
		SELECT id, brand, model, on_hand FROM spare_parts WHERE id = ?
	`, id).Scan(&part.ID, &part.Brand, &part.Model, &part.OnHand)
	if errors.Is(err, sql.ErrNoRows) {
		return SparePart{}, ErrNotFound
	}
	if err != nil {
		return SparePart{}, fmt.Errorf("read spare part: %w", err)
	}
	return part, nil
}

func findSpare(ctx context.Context, q spareDB, brand string, model string) (SparePart, error) {
	var part SparePart
	err := q.QueryRowContext(ctx, `
		SELECT id, brand, model, on_hand FROM spare_parts WHERE brand_fold = ? AND model_fold = ?
	`, fold(brand), fold(model)).Scan(&part.ID, &part.Brand, &part.Model, &part.OnHand)
	if errors.Is(err, sql.ErrNoRows) {
		return SparePart{}, ErrNotFound
	}
	if err != nil {
		return SparePart{}, fmt.Errorf("find spare part: %w", err)
	}
	return part, nil
}

func loadServiceDevice(ctx context.Context, q spareDB, id string) (serviceDevice, error) {
	var item serviceDevice
	err := q.QueryRowContext(ctx, `
		SELECT services.status, devices.brand, devices.model, devices.serial
		FROM services
		JOIN devices ON devices.id = services.device_id
		WHERE services.id = ?
	`, id).Scan(&item.Status, &item.Brand, &item.Model, &item.Serial)
	if errors.Is(err, sql.ErrNoRows) {
		return serviceDevice{}, ErrNotFound
	}
	if err != nil {
		return serviceDevice{}, fmt.Errorf("read service: %w", err)
	}
	return item, nil
}

func requireService(ctx context.Context, database *sql.DB, id string) error {
	var one int
	err := database.QueryRowContext(ctx, `SELECT 1 FROM services WHERE id = ?`, id).Scan(&one)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	if err != nil {
		return fmt.Errorf("read service: %w", err)
	}
	return nil
}

func spareNames(brand string, model string, fields *FieldError) (string, string) {
	return requiredText(brand, maxNameLen, "brand", fields), requiredText(model, maxNameLen, "model", fields)
}

func quantityIn(value *float64, min int, max int, fields *FieldError) int {
	if value == nil {
		fields.addMissing("quantity")
		return 0
	}
	number := *value
	if number < float64(min) || number > float64(max) || number != float64(int64(number)) {
		fields.addInvalid("quantity")
		return 0
	}
	return int(number)
}

func spareIdentityConflict(err error) bool {
	return err != nil && strings.Contains(err.Error(), "UNIQUE") && strings.Contains(err.Error(), "spare_parts_identity")
}

func scanNames(rows *sql.Rows) ([]string, error) {
	var names []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, fmt.Errorf("scan spare name: %w", err)
		}
		names = append(names, name)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list spare names: %w", err)
	}
	if names == nil {
		return []string{}, nil
	}
	return names, nil
}

func pageCut[T any](items []T, limit int) ([]T, bool) {
	if len(items) > limit {
		return items[:limit], true
	}
	if items == nil {
		return []T{}, false
	}
	return items, false
}
