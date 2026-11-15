package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

const maxStateName = 80

const (
	EventStateAdded      = "state.added"
	EventStateUpdated    = "state.updated"
	EventStatesReordered = "states.reordered"
)

type stateSeed struct {
	Code   string
	NameEn string
	NameTr string
}

func fixedStateSeeds() []stateSeed {
	return []stateSeed{
		{StatusRegistered, "Registered", "Kayıt yapıldı"},
		{StatusRepairWarranty, "Repair under warranty", "Garanti kapsamında tamiri yapılıyor"},
		{StatusRepairOutOfWarranty, "Repair outside warranty", "Garanti dışında tamiri yapılıyor"},
		{StatusCancelled, "Cancelled", "İptal edildi"},
		{StatusDelivered, "Delivered", "Teslim edildi"},
	}
}

func legacyStateSeeds() []stateSeed {
	return []stateSeed{
		{StatusInspecting, "Inspecting", "Ön inceleme yapılıyor"},
		{StatusWaitingPartnerPart, "Waiting for a partner part", "İş ortağından yedek parça bekleniyor"},
		{StatusWaitingOwnPart, "Waiting for a part we will buy", "Kendi satın alacağımız yedek parça bekleniyor"},
		{StatusFinalTesting, "Final tests", "Final testleri yapılıyor"},
		{StatusPreparingDelivery, "Preparing for delivery", "Teslim için hazırlanıyor"},
		{StatusInStorage, "In storage", "Depoda"},
		{StatusLeftStorage, "Left storage", "Depodan çıkış yapıldı"},
	}
}

func lockedState(code string) bool {
	return code == StatusRegistered || code == StatusDelivered || code == StatusCancelled
}

func fixedState(code string) bool {
	for _, item := range fixedStateSeeds() {
		if item.Code == code {
			return true
		}
	}
	return false
}

type ServiceState struct {
	Code       string
	NameEn     string
	NameTr     string
	Position   int
	Active     bool
	ShowOnHome bool
	Fixed      bool
	Locked     bool
	Used       bool
}

type StateInput struct {
	NameEn     string
	NameTr     string
	ShowOnHome bool
}

type StatePatch struct {
	NameEn     *string
	NameTr     *string
	Active     *bool
	ShowOnHome *bool
}

func seedServiceStates(ctx context.Context, tx *sql.Tx) error {
	position := 1
	for _, item := range fixedStateSeeds() {
		if err := insertState(ctx, tx, item, position, true, false); err != nil {
			return err
		}
		position++
	}
	present, current, err := storedStatusCodes(ctx, tx)
	if err != nil {
		return err
	}
	known := map[string]bool{}
	for _, item := range fixedStateSeeds() {
		known[item.Code] = true
	}
	for _, item := range legacyStateSeeds() {
		known[item.Code] = true
		if !present[item.Code] {
			continue
		}
		if err := insertState(ctx, tx, item, position, current[item.Code], false); err != nil {
			return err
		}
		position++
	}
	extras := make([]string, 0)
	for code := range present {
		if !known[code] {
			extras = append(extras, code)
		}
	}
	sortStrings(extras)
	for _, code := range extras {
		item := stateSeed{Code: code, NameEn: code, NameTr: code}
		if err := insertState(ctx, tx, item, position, current[code], false); err != nil {
			return err
		}
		position++
	}
	return nil
}

func sortStrings(values []string) {
	for i := 1; i < len(values); i++ {
		for j := i; j > 0 && values[j] < values[j-1]; j-- {
			values[j], values[j-1] = values[j-1], values[j]
		}
	}
}

func storedStatusCodes(ctx context.Context, tx *sql.Tx) (map[string]bool, map[string]bool, error) {
	present := map[string]bool{}
	current := map[string]bool{}
	rows, err := tx.QueryContext(ctx, `SELECT status FROM services`)
	if err != nil {
		return nil, nil, fmt.Errorf("migrate statuses: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var code string
		if err := rows.Scan(&code); err != nil {
			return nil, nil, fmt.Errorf("migrate status: %w", err)
		}
		present[code] = true
		current[code] = true
	}
	if err := rows.Err(); err != nil {
		return nil, nil, fmt.Errorf("migrate statuses: %w", err)
	}
	history, err := tx.QueryContext(ctx, `SELECT from_status, to_status FROM service_status_changes`)
	if err != nil {
		return nil, nil, fmt.Errorf("migrate history: %w", err)
	}
	defer history.Close()
	for history.Next() {
		var from, to string
		if err := history.Scan(&from, &to); err != nil {
			return nil, nil, fmt.Errorf("migrate history: %w", err)
		}
		present[from] = true
		present[to] = true
	}
	if err := history.Err(); err != nil {
		return nil, nil, fmt.Errorf("migrate history: %w", err)
	}
	return present, current, nil
}

func insertState(ctx context.Context, tx *sql.Tx, item stateSeed, position int, active bool, show bool) error {
	activeValue := 0
	if active {
		activeValue = 1
	}
	showValue := 0
	if show {
		showValue = 1
	}
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO service_states (code, name_en, name_tr, name_en_fold, name_tr_fold, position, active, show_on_home)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, item.Code, item.NameEn, item.NameTr, fold(item.NameEn), fold(item.NameTr), position, activeValue, showValue); err != nil {
		return fmt.Errorf("insert state %s: %w", item.Code, err)
	}
	return nil
}

func ListStates(ctx context.Context, database *sql.DB) ([]ServiceState, error) {
	rows, err := database.QueryContext(ctx, `
		SELECT code, name_en, name_tr, position, active, show_on_home,
			EXISTS(SELECT 1 FROM services WHERE services.status = service_states.code)
		FROM service_states
		ORDER BY position, code
	`)
	if err != nil {
		return nil, fmt.Errorf("list states: %w", err)
	}
	defer rows.Close()
	items := []ServiceState{}
	for rows.Next() {
		item, err := scanState(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list states: %w", err)
	}
	return items, nil
}

type stateScanner interface {
	Scan(dest ...any) error
}

func scanState(row stateScanner) (ServiceState, error) {
	var item ServiceState
	var active, show, used int
	if err := row.Scan(&item.Code, &item.NameEn, &item.NameTr, &item.Position, &active, &show, &used); err != nil {
		return ServiceState{}, fmt.Errorf("scan state: %w", err)
	}
	item.Active = active == 1
	item.ShowOnHome = show == 1
	item.Used = used == 1
	item.Fixed = fixedState(item.Code)
	item.Locked = lockedState(item.Code)
	return item, nil
}

func AddState(ctx context.Context, database *sql.DB, actor User, input StateInput) (ServiceState, error) {
	fields := &FieldError{}
	nameEn := cleanStateName(input.NameEn, "nameEn", fields)
	nameTr := cleanStateName(input.NameTr, "nameTr", fields)
	if err := fields.asError(); err != nil {
		return ServiceState{}, err
	}
	code := stateCode(nameEn)
	if code == "" {
		fields.addTaken("nameEn")
		return ServiceState{}, fields
	}
	tx, err := database.BeginTx(ctx, nil)
	if err != nil {
		return ServiceState{}, fmt.Errorf("begin state: %w", err)
	}
	defer tx.Rollback()
	if err := rejectTakenNames(ctx, tx, nameEn, nameTr, "", fields); err != nil {
		return ServiceState{}, err
	}
	if err := fields.asError(); err != nil {
		return ServiceState{}, err
	}
	var existing int
	if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM service_states WHERE code = ?`, code).Scan(&existing); err != nil {
		return ServiceState{}, fmt.Errorf("state code: %w", err)
	}
	if existing != 0 {
		fields.addTaken("nameEn")
		return ServiceState{}, fields
	}
	var position int
	if err := tx.QueryRowContext(ctx, `SELECT COALESCE(MAX(position), 0) + 1 FROM service_states`).Scan(&position); err != nil {
		return ServiceState{}, fmt.Errorf("state position: %w", err)
	}
	show := 0
	if input.ShowOnHome {
		show = 1
	}
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO service_states (code, name_en, name_tr, name_en_fold, name_tr_fold, position, active, show_on_home)
		VALUES (?, ?, ?, ?, ?, ?, 1, ?)
	`, code, nameEn, nameTr, fold(nameEn), fold(nameTr), position, show); err != nil {
		return ServiceState{}, fmt.Errorf("insert state: %w", err)
	}
	detail := map[string]string{"nameEn": nameEn, "nameTr": nameTr}
	if input.ShowOnHome {
		detail["showOnHome"] = "shown"
	}
	if err := insertAudit(ctx, tx, actor, EventStateAdded, nameEn+" "+nameTr, detail); err != nil {
		return ServiceState{}, err
	}
	if err := tx.Commit(); err != nil {
		return ServiceState{}, fmt.Errorf("commit state: %w", err)
	}
	return loadState(ctx, database, code)
}

func UpdateState(ctx context.Context, database *sql.DB, actor User, code string, patch StatePatch) (ServiceState, error) {
	tx, err := database.BeginTx(ctx, nil)
	if err != nil {
		return ServiceState{}, fmt.Errorf("begin state: %w", err)
	}
	defer tx.Rollback()
	current, err := loadState(ctx, tx, code)
	if err != nil {
		return ServiceState{}, err
	}
	fields := &FieldError{}
	next := current
	renamed := false
	if patch.NameEn != nil || patch.NameTr != nil {
		if patch.NameEn == nil || patch.NameTr == nil {
			if patch.NameEn == nil {
				fields.addMissing("nameEn")
			}
			if patch.NameTr == nil {
				fields.addMissing("nameTr")
			}
			return ServiceState{}, fields
		}
		next.NameEn = cleanStateName(*patch.NameEn, "nameEn", fields)
		next.NameTr = cleanStateName(*patch.NameTr, "nameTr", fields)
		if err := fields.asError(); err != nil {
			return ServiceState{}, err
		}
		if current.Fixed && (next.NameEn != current.NameEn || next.NameTr != current.NameTr) {
			fields.addInvalid("nameEn")
			return ServiceState{}, fields
		}
		if err := rejectTakenNames(ctx, tx, next.NameEn, next.NameTr, code, fields); err != nil {
			return ServiceState{}, err
		}
		if err := fields.asError(); err != nil {
			return ServiceState{}, err
		}
		renamed = next.NameEn != current.NameEn || next.NameTr != current.NameTr
	}
	activeChanged := false
	if patch.Active != nil && *patch.Active != current.Active {
		if current.Locked && !*patch.Active {
			fields.addInvalid("active")
			return ServiceState{}, fields
		}
		next.Active = *patch.Active
		activeChanged = true
	}
	homeChanged := false
	if patch.ShowOnHome != nil && *patch.ShowOnHome != current.ShowOnHome {
		if current.Locked && *patch.ShowOnHome {
			fields.addInvalid("showOnHome")
			return ServiceState{}, fields
		}
		next.ShowOnHome = *patch.ShowOnHome
		homeChanged = true
	}
	if !renamed && !activeChanged && !homeChanged {
		return current, nil
	}
	activeValue := 0
	if next.Active {
		activeValue = 1
	}
	showValue := 0
	if next.ShowOnHome {
		showValue = 1
	}
	if _, err := tx.ExecContext(ctx, `
		UPDATE service_states
		SET name_en = ?, name_tr = ?, name_en_fold = ?, name_tr_fold = ?, active = ?, show_on_home = ?
		WHERE code = ?
	`, next.NameEn, next.NameTr, fold(next.NameEn), fold(next.NameTr), activeValue, showValue, code); err != nil {
		return ServiceState{}, fmt.Errorf("update state: %w", err)
	}
	detail := map[string]string{"nameEn": next.NameEn, "nameTr": next.NameTr}
	if renamed {
		detail["previousNameEn"] = current.NameEn
		detail["previousNameTr"] = current.NameTr
	}
	if activeChanged {
		detail["active"] = flagWord(next.Active, "active", "inactive")
		detail["previousActive"] = flagWord(current.Active, "active", "inactive")
	}
	if homeChanged {
		detail["home"] = flagWord(next.ShowOnHome, "shown", "hidden")
		detail["previousHome"] = flagWord(current.ShowOnHome, "shown", "hidden")
	}
	subject := strings.TrimSpace(current.NameEn + " " + next.NameEn + " " + current.NameTr + " " + next.NameTr)
	if err := insertAudit(ctx, tx, actor, EventStateUpdated, subject, detail); err != nil {
		return ServiceState{}, err
	}
	if err := tx.Commit(); err != nil {
		return ServiceState{}, fmt.Errorf("commit state: %w", err)
	}
	return loadState(ctx, database, code)
}

func MoveState(ctx context.Context, database *sql.DB, actor User, code string, position int) (ServiceState, error) {
	tx, err := database.BeginTx(ctx, nil)
	if err != nil {
		return ServiceState{}, fmt.Errorf("begin state move: %w", err)
	}
	defer tx.Rollback()
	rows, err := tx.QueryContext(ctx, `SELECT code FROM service_states ORDER BY position`)
	if err != nil {
		return ServiceState{}, fmt.Errorf("state order: %w", err)
	}
	var codes []string
	for rows.Next() {
		var item string
		if err := rows.Scan(&item); err != nil {
			rows.Close()
			return ServiceState{}, fmt.Errorf("state order: %w", err)
		}
		codes = append(codes, item)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return ServiceState{}, fmt.Errorf("state order: %w", err)
	}
	rows.Close()
	from := -1
	for i, item := range codes {
		if item == code {
			from = i
			break
		}
	}
	if from < 0 {
		return ServiceState{}, ErrNotFound
	}
	if position < 1 || position > len(codes) {
		return ServiceState{}, &FieldError{Invalid: []string{"position"}}
	}
	to := position - 1
	if from == to {
		return loadState(ctx, tx, code)
	}
	movedCode := codes[from]
	next := make([]string, 0, len(codes))
	next = append(next, codes[:from]...)
	next = append(next, codes[from+1:]...)
	codes = make([]string, 0, len(next)+1)
	codes = append(codes, next[:to]...)
	codes = append(codes, movedCode)
	codes = append(codes, next[to:]...)
	for i, item := range codes {
		if _, err := tx.ExecContext(ctx, `UPDATE service_states SET position = ? WHERE code = ?`, -(i + 1), item); err != nil {
			return ServiceState{}, fmt.Errorf("move state: %w", err)
		}
	}
	for i, item := range codes {
		if _, err := tx.ExecContext(ctx, `UPDATE service_states SET position = ? WHERE code = ?`, i+1, item); err != nil {
			return ServiceState{}, fmt.Errorf("move state: %w", err)
		}
	}
	moved, err := loadState(ctx, tx, code)
	if err != nil {
		return ServiceState{}, err
	}
	if err := insertAudit(ctx, tx, actor, EventStatesReordered, moved.NameEn+" "+moved.NameTr, map[string]string{
		"nameEn": moved.NameEn,
		"nameTr": moved.NameTr,
	}); err != nil {
		return ServiceState{}, err
	}
	if err := tx.Commit(); err != nil {
		return ServiceState{}, fmt.Errorf("commit state move: %w", err)
	}
	return moved, nil
}

func loadState(ctx context.Context, q queryer, code string) (ServiceState, error) {
	row := q.QueryRowContext(ctx, `
		SELECT code, name_en, name_tr, position, active, show_on_home,
			EXISTS(SELECT 1 FROM services WHERE services.status = service_states.code)
		FROM service_states
		WHERE code = ?
	`, code)
	item, err := scanState(row)
	if errors.Is(err, sql.ErrNoRows) {
		return ServiceState{}, ErrNotFound
	}
	return item, err
}

func activeStatus(ctx context.Context, q queryer, code string) (bool, error) {
	var active int
	err := q.QueryRowContext(ctx, `SELECT active FROM service_states WHERE code = ?`, code).Scan(&active)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("state active: %w", err)
	}
	return active == 1, nil
}

func filterableStatus(ctx context.Context, q queryer, code string) error {
	var active, used int
	err := q.QueryRowContext(ctx, `
		SELECT active, EXISTS(SELECT 1 FROM services WHERE status = service_states.code)
		FROM service_states
		WHERE code = ?
	`, code).Scan(&active, &used)
	if err == sql.ErrNoRows || (err == nil && active == 0 && used == 0) {
		return ErrInvalidStatus
	}
	if err != nil {
		return fmt.Errorf("state filter: %w", err)
	}
	return nil
}

func cleanStateName(value string, field string, fields *FieldError) string {
	value = strings.TrimSpace(value)
	if value == "" {
		fields.addMissing(field)
		return ""
	}
	if len([]rune(value)) > maxStateName {
		fields.addTooLong(field)
	}
	return value
}

func rejectTakenNames(ctx context.Context, q queryer, nameEn string, nameTr string, except string, fields *FieldError) error {
	enTaken, err := foldedNameTaken(ctx, q, "name_en_fold", fold(nameEn), except)
	if err != nil {
		return err
	}
	trTaken, err := foldedNameTaken(ctx, q, "name_tr_fold", fold(nameTr), except)
	if err != nil {
		return err
	}
	if enTaken {
		fields.addTaken("nameEn")
	}
	if trTaken {
		fields.addTaken("nameTr")
	}
	return nil
}

func foldedNameTaken(ctx context.Context, q queryer, column string, folded string, except string) (bool, error) {
	var code string
	err := q.QueryRowContext(ctx, `SELECT code FROM service_states WHERE `+column+` = ? AND code != ?`, folded, except).Scan(&code)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("state name: %w", err)
	}
	return true, nil
}

func stateCode(english string) string {
	var b strings.Builder
	underscore := false
	for _, r := range fold(english) {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
			underscore = false
			continue
		}
		if b.Len() > 0 && !underscore {
			b.WriteByte('_')
			underscore = true
		}
	}
	return strings.Trim(b.String(), "_")
}

func flagWord(value bool, yes string, no string) string {
	if value {
		return yes
	}
	return no
}

func (e *FieldError) addTaken(field string) {
	e.Taken = append(e.Taken, field)
}
