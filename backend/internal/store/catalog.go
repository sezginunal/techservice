package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"techservice/internal/catalog"
)

const (
	catalogFile  = "catalog"
	catalogAdded = "added"
)

type CatalogModel struct {
	ID     string
	Brand  string
	Model  string
	Source string
	Active bool
}

type CatalogPair struct {
	Brand string
	Model string
}

type CatalogListQuery struct {
	Text       string
	Limit      int
	AfterBrand string
	AfterModel string
	AfterID    string
	HasAfter   bool
}

type CatalogPage struct {
	Items   []CatalogModel
	HasMore bool
}

func CatalogKey(brand, model string) string {
	return fold(brand) + "\x00" + fold(model)
}

func MergeCatalog(ctx context.Context, database *sql.DB, brands []catalog.Brand) error {
	tx, err := database.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin catalog merge: %w", err)
	}
	defer tx.Rollback()

	seen := map[string]bool{}
	for _, item := range brands {
		for _, model := range item.Models {
			key := CatalogKey(item.Brand, model)
			if seen[key] {
				continue
			}
			seen[key] = true
			var id string
			err := tx.QueryRowContext(ctx, `
				SELECT id FROM catalog_models WHERE brand_fold = ? AND model_fold = ?
			`, fold(item.Brand), fold(model)).Scan(&id)
			if err == nil {
				continue
			}
			if !errors.Is(err, sql.ErrNoRows) {
				return fmt.Errorf("find catalog model: %w", err)
			}
			if err := insertCatalogModel(ctx, tx, nil, item.Brand, model, catalogFile); err != nil {
				if catalogConflict(err) {
					continue
				}
				return err
			}
		}
	}

	rows, err := tx.QueryContext(ctx, `
		SELECT id, brand_fold, model_fold, active FROM catalog_models WHERE source = ?
	`, catalogFile)
	if err != nil {
		return fmt.Errorf("list catalog models: %w", err)
	}
	defer rows.Close()
	type stored struct {
		id     string
		key    string
		active bool
	}
	var current []stored
	for rows.Next() {
		var item stored
		var brandFold, modelFold string
		var active int
		if err := rows.Scan(&item.id, &brandFold, &modelFold, &active); err != nil {
			return fmt.Errorf("scan catalog model: %w", err)
		}
		item.key = brandFold + "\x00" + modelFold
		item.active = active == 1
		current = append(current, item)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("list catalog models: %w", err)
	}
	if err := rows.Close(); err != nil {
		return fmt.Errorf("list catalog models: %w", err)
	}
	for _, item := range current {
		if seen[item.key] || !item.active {
			continue
		}
		if _, err := tx.ExecContext(ctx, `UPDATE catalog_models SET active = 0 WHERE id = ? AND active = 1`, item.id); err != nil {
			return fmt.Errorf("deactivate catalog model: %w", err)
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit catalog merge: %w", err)
	}
	return nil
}

func ListCatalogModels(ctx context.Context, database *sql.DB, query CatalogListQuery) (CatalogPage, error) {
	limit, err := pageLimit(query.Limit)
	if err != nil {
		return CatalogPage{}, err
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
		SELECT id, brand, model, source, active
		FROM catalog_models
		`+where+`
		ORDER BY brand COLLATE NOCASE, model COLLATE NOCASE, id
		LIMIT ?
	`, args...)
	if err != nil {
		return CatalogPage{}, fmt.Errorf("list catalog models: %w", err)
	}
	defer rows.Close()
	var items []CatalogModel
	for rows.Next() {
		item, err := scanCatalogModel(rows)
		if err != nil {
			return CatalogPage{}, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return CatalogPage{}, fmt.Errorf("list catalog models: %w", err)
	}
	items, more := pageCut(items, limit)
	if items == nil {
		items = []CatalogModel{}
	}
	return CatalogPage{Items: items, HasMore: more}, nil
}

func CreateCatalogModel(ctx context.Context, database *sql.DB, actor User, brand, model string) (CatalogModel, error) {
	fields := &FieldError{}
	brand, model = catalogNames(brand, model, fields)
	if err := fields.asError(); err != nil {
		return CatalogModel{}, err
	}
	tx, err := database.BeginTx(ctx, nil)
	if err != nil {
		return CatalogModel{}, fmt.Errorf("begin catalog model: %w", err)
	}
	defer tx.Rollback()
	if _, err := findCatalogModel(ctx, tx, brand, model); err == nil {
		return CatalogModel{}, &FieldError{Taken: []string{"catalog"}}
	} else if !errors.Is(err, ErrNotFound) {
		return CatalogModel{}, err
	}
	if err := insertCatalogModel(ctx, tx, &actor, brand, model, catalogAdded); err != nil {
		if catalogConflict(err) {
			return CatalogModel{}, &FieldError{Taken: []string{"catalog"}}
		}
		return CatalogModel{}, err
	}
	if err := tx.Commit(); err != nil {
		return CatalogModel{}, fmt.Errorf("commit catalog model: %w", err)
	}
	return findCatalogModel(ctx, database, brand, model)
}

func SetCatalogActive(ctx context.Context, database *sql.DB, actor User, id string, active bool) (CatalogModel, error) {
	tx, err := database.BeginTx(ctx, nil)
	if err != nil {
		return CatalogModel{}, fmt.Errorf("begin catalog model: %w", err)
	}
	defer tx.Rollback()
	current, err := loadCatalogModel(ctx, tx, id)
	if err != nil {
		return CatalogModel{}, err
	}
	if current.Active == active {
		return current, nil
	}
	value := 0
	if active {
		value = 1
	}
	if _, err := tx.ExecContext(ctx, `UPDATE catalog_models SET active = ? WHERE id = ?`, value, id); err != nil {
		return CatalogModel{}, fmt.Errorf("update catalog model: %w", err)
	}
	if err := insertAudit(ctx, tx, actor, EventModelUpdated, current.Brand+" "+current.Model, map[string]string{
		"brand":          current.Brand,
		"model":          current.Model,
		"active":         flagWord(active, "active", "inactive"),
		"previousActive": flagWord(current.Active, "active", "inactive"),
	}); err != nil {
		return CatalogModel{}, err
	}
	if err := tx.Commit(); err != nil {
		return CatalogModel{}, fmt.Errorf("commit catalog model: %w", err)
	}
	return loadCatalogModel(ctx, database, id)
}

func SuggestCatalogBrands(ctx context.Context, database *sql.DB, text string, activeOnly bool) ([]string, error) {
	where := "brand_fold LIKE ? ESCAPE '\\'"
	args := []any{likePattern(text)}
	if activeOnly {
		where = "active = 1 AND " + where
	}
	rows, err := database.QueryContext(ctx, `
		SELECT MIN(brand)
		FROM catalog_models
		WHERE `+where+`
		GROUP BY brand_fold
		ORDER BY MIN(brand) COLLATE NOCASE
		LIMIT ?
	`, append(args, suggestLimit)...)
	if err != nil {
		return nil, fmt.Errorf("suggest catalog brands: %w", err)
	}
	defer rows.Close()
	return scanNames(rows)
}

func SuggestCatalogModels(ctx context.Context, database *sql.DB, brand, text string, activeOnly bool) ([]CatalogPair, error) {
	if strings.TrimSpace(brand) == "" {
		return []CatalogPair{}, nil
	}
	where := "brand_fold = ? AND model_fold LIKE ? ESCAPE '\\'"
	args := []any{fold(brand), likePattern(text)}
	if activeOnly {
		where = "active = 1 AND " + where
	}
	rows, err := database.QueryContext(ctx, `
		SELECT brand, model
		FROM catalog_models
		WHERE `+where+`
		ORDER BY model COLLATE NOCASE
		LIMIT ?
	`, append(args, suggestLimit)...)
	if err != nil {
		return nil, fmt.Errorf("suggest catalog models: %w", err)
	}
	defer rows.Close()
	var items []CatalogPair
	for rows.Next() {
		var item CatalogPair
		if err := rows.Scan(&item.Brand, &item.Model); err != nil {
			return nil, fmt.Errorf("scan catalog model: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("suggest catalog models: %w", err)
	}
	if items == nil {
		items = []CatalogPair{}
	}
	return items, nil
}

func CatalogCanon(ctx context.Context, database *sql.DB, pairs []CatalogPair) (map[string]CatalogPair, error) {
	unique := map[string]CatalogPair{}
	for _, pair := range pairs {
		if fold(pair.Brand) == "" || fold(pair.Model) == "" {
			continue
		}
		unique[CatalogKey(pair.Brand, pair.Model)] = pair
	}
	if len(unique) == 0 {
		return map[string]CatalogPair{}, nil
	}
	clauses := make([]string, 0, len(unique))
	args := make([]any, 0, len(unique)*2)
	for _, pair := range unique {
		clauses = append(clauses, "(brand_fold = ? AND model_fold = ?)")
		args = append(args, fold(pair.Brand), fold(pair.Model))
	}
	rows, err := database.QueryContext(ctx, `
		SELECT brand, model FROM catalog_models WHERE `+strings.Join(clauses, " OR "), args...)
	if err != nil {
		return nil, fmt.Errorf("catalog names: %w", err)
	}
	defer rows.Close()
	found := map[string]CatalogPair{}
	for rows.Next() {
		var item CatalogPair
		if err := rows.Scan(&item.Brand, &item.Model); err != nil {
			return nil, fmt.Errorf("catalog names: %w", err)
		}
		found[CatalogKey(item.Brand, item.Model)] = item
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("catalog names: %w", err)
	}
	return found, nil
}

func ensureCatalogModel(ctx context.Context, tx *sql.Tx, actor User, brand, model string) error {
	if _, err := findCatalogModel(ctx, tx, brand, model); err == nil {
		return nil
	} else if !errors.Is(err, ErrNotFound) {
		return err
	}
	if err := insertCatalogModel(ctx, tx, &actor, brand, model, catalogAdded); err != nil {
		if catalogConflict(err) {
			return nil
		}
		return err
	}
	return nil
}

func insertCatalogModel(ctx context.Context, tx *sql.Tx, actor *User, brand, model, source string) error {
	id, err := newID()
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `
		INSERT INTO catalog_models (id, brand, model, brand_fold, model_fold, source, active, created_at)
		VALUES (?, ?, ?, ?, ?, ?, 1, ?)
	`, id, brand, model, fold(brand), fold(model), source, timestamp(time.Now()))
	if err != nil {
		return err
	}
	if actor == nil {
		return nil
	}
	return insertAudit(ctx, tx, *actor, EventModelAdded, brand+" "+model, map[string]string{
		"brand": brand,
		"model": model,
	})
}

func findCatalogModel(ctx context.Context, q spareDB, brand, model string) (CatalogModel, error) {
	row := q.QueryRowContext(ctx, `
		SELECT id, brand, model, source, active FROM catalog_models WHERE brand_fold = ? AND model_fold = ?
	`, fold(brand), fold(model))
	item, err := scanCatalogRow(row)
	if errors.Is(err, sql.ErrNoRows) {
		return CatalogModel{}, ErrNotFound
	}
	if err != nil {
		return CatalogModel{}, fmt.Errorf("find catalog model: %w", err)
	}
	return item, nil
}

func loadCatalogModel(ctx context.Context, q spareDB, id string) (CatalogModel, error) {
	row := q.QueryRowContext(ctx, `
		SELECT id, brand, model, source, active FROM catalog_models WHERE id = ?
	`, id)
	item, err := scanCatalogRow(row)
	if errors.Is(err, sql.ErrNoRows) {
		return CatalogModel{}, ErrNotFound
	}
	if err != nil {
		return CatalogModel{}, fmt.Errorf("read catalog model: %w", err)
	}
	return item, nil
}

type catalogScanner interface {
	Scan(dest ...any) error
}

func scanCatalogModel(rows *sql.Rows) (CatalogModel, error) {
	item, err := scanCatalogRow(rows)
	if err != nil {
		return CatalogModel{}, fmt.Errorf("scan catalog model: %w", err)
	}
	return item, nil
}

func scanCatalogRow(row catalogScanner) (CatalogModel, error) {
	var item CatalogModel
	var active int
	if err := row.Scan(&item.ID, &item.Brand, &item.Model, &item.Source, &active); err != nil {
		return CatalogModel{}, err
	}
	item.Active = active == 1
	return item, nil
}

func catalogNames(brand, model string, fields *FieldError) (string, string) {
	return requiredText(brand, maxNameLen, "brand", fields), requiredText(model, maxNameLen, "model", fields)
}

func catalogConflict(err error) bool {
	return err != nil && strings.Contains(err.Error(), "UNIQUE") && strings.Contains(err.Error(), "catalog_models")
}
