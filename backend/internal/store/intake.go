package store

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

type Customer struct {
	ID        int64
	Name      string
	Phone     string
	Email     string
	CreatedAt time.Time
}

type Photo struct {
	ID        string
	MediaType string
	Size      int64
	Path      string
}

type Part struct {
	ID     string
	Name   string
	Note   string
	Photos []Photo
}

type Device struct {
	ID        string
	Customer  Customer
	Name      string
	Brand     string
	Model     string
	Serial    string
	Note      string
	Status    string
	CreatedAt time.Time
	Parts     []Part
	Photos    []Photo
}

type DeviceListItem struct {
	ID           string
	Name         string
	Brand        string
	Model        string
	CustomerName string
	Status       string
	CreatedAt    time.Time
}

type Upload struct {
	Open func() (io.ReadCloser, error)
}

type NewPart struct {
	Name   string
	Note   string
	Photos []int
}

type NewDevice struct {
	CustomerID    *int64
	CustomerName  string
	CustomerPhone string
	CustomerEmail string
	Name          string
	Brand         string
	Model         string
	Serial        string
	Note          string
	Parts         []NewPart
	Files         []Upload
	DevicePhotos  []int
}

type PartChange struct {
	ID     string
	Name   string
	Note   string
	Keep   []string
	Photos []int
}

type DeviceChange struct {
	Name          string
	Brand         string
	Model         string
	Serial        string
	Note          string
	CustomerName  string
	CustomerPhone string
	CustomerEmail string
	Parts         []PartChange
	Files         []Upload
	DeviceKeep    []string
	DevicePhotos  []int
}

type CustomerQuery struct {
	Text      string
	Limit     int
	AfterName string
	AfterID   int64
	HasAfter  bool
}

type CustomerPage struct {
	Items   []Customer
	HasMore bool
}

type DeviceQuery struct {
	Text        string
	CustomerID  int64
	HasCustomer bool
	Limit       int
	AfterTime   string
	AfterID     string
	HasAfter    bool
}

type DevicePage struct {
	Items   []DeviceListItem
	HasMore bool
}

func ListCustomers(ctx context.Context, database *sql.DB, query CustomerQuery) (CustomerPage, error) {
	limit, err := pageLimit(query.Limit)
	if err != nil {
		return CustomerPage{}, err
	}
	clauses := []string{}
	args := []any{}
	if strings.TrimSpace(query.Text) != "" {
		pattern := likePattern(query.Text)
		clauses = append(clauses, "(name_fold LIKE ? ESCAPE '\\' OR phone_fold LIKE ? ESCAPE '\\' OR email_fold LIKE ? ESCAPE '\\')")
		args = append(args, pattern, pattern, pattern)
	}
	if query.HasAfter {
		clauses = append(clauses, "(name COLLATE NOCASE > ? OR (name COLLATE NOCASE = ? AND id > ?))")
		args = append(args, query.AfterName, query.AfterName, query.AfterID)
	}
	where := ""
	if len(clauses) > 0 {
		where = "WHERE " + strings.Join(clauses, " AND ")
	}
	args = append(args, limit+1)
	rows, err := database.QueryContext(ctx, `
		SELECT id, name, phone, email, created_at
		FROM customers
		`+where+`
		ORDER BY name COLLATE NOCASE, id
		LIMIT ?
	`, args...)
	if err != nil {
		return CustomerPage{}, fmt.Errorf("list customers: %w", err)
	}
	defer rows.Close()

	var items []Customer
	for rows.Next() {
		customer, err := scanCustomer(rows)
		if err != nil {
			return CustomerPage{}, err
		}
		items = append(items, customer)
	}
	if err := rows.Err(); err != nil {
		return CustomerPage{}, fmt.Errorf("list customers: %w", err)
	}
	page := CustomerPage{Items: items}
	if len(items) > limit {
		page.HasMore = true
		page.Items = items[:limit]
	}
	if page.Items == nil {
		page.Items = []Customer{}
	}
	return page, nil
}

func GetCustomer(ctx context.Context, database *sql.DB, id int64) (Customer, error) {
	return loadCustomer(ctx, database, id)
}

func UpdateCustomer(ctx context.Context, database *sql.DB, actor User, id int64, name string, phone string, email string) (Customer, error) {
	fields := &FieldError{}
	name = requiredText(name, maxNameLen, "customerName", fields)
	phone = optionalText(phone, maxPhoneLen, "phone", fields)
	email = optionalEmail(email, fields)
	if err := fields.asError(); err != nil {
		return Customer{}, err
	}

	tx, err := database.BeginTx(ctx, nil)
	if err != nil {
		return Customer{}, fmt.Errorf("begin customer: %w", err)
	}
	defer tx.Rollback()

	current, err := loadCustomer(ctx, tx, id)
	if err != nil {
		return Customer{}, err
	}
	next, err := writeCustomer(ctx, tx, actor, current, name, phone, email)
	if err != nil {
		return Customer{}, err
	}
	if err := tx.Commit(); err != nil {
		return Customer{}, fmt.Errorf("commit customer: %w", err)
	}
	return next, nil
}

func ListDevices(ctx context.Context, database *sql.DB, query DeviceQuery) (DevicePage, error) {
	limit, err := pageLimit(query.Limit)
	if err != nil {
		return DevicePage{}, err
	}
	clauses := []string{}
	args := []any{}
	if query.HasCustomer {
		clauses = append(clauses, "devices.customer_id = ?")
		args = append(args, query.CustomerID)
	}
	if strings.TrimSpace(query.Text) != "" {
		pattern := likePattern(query.Text)
		clauses = append(clauses, `(
			devices.name_fold LIKE ? ESCAPE '\' OR devices.brand_fold LIKE ? ESCAPE '\'
			OR devices.model_fold LIKE ? ESCAPE '\' OR devices.serial_fold LIKE ? ESCAPE '\'
			OR devices.id LIKE ? ESCAPE '\' OR customers.name_fold LIKE ? ESCAPE '\'
			OR customers.phone_fold LIKE ? ESCAPE '\'
		)`)
		args = append(args, pattern, pattern, pattern, pattern, pattern, pattern, pattern)
	}
	if query.HasAfter {
		clauses = append(clauses, "(devices.created_at < ? OR (devices.created_at = ? AND devices.id < ?))")
		args = append(args, query.AfterTime, query.AfterTime, query.AfterID)
	}
	where := ""
	if len(clauses) > 0 {
		where = "WHERE " + strings.Join(clauses, " AND ")
	}
	args = append(args, limit+1)
	rows, err := database.QueryContext(ctx, `
		SELECT devices.id, devices.name, devices.brand, devices.model, customers.name, devices.status, devices.created_at
		FROM devices
		JOIN customers ON customers.id = devices.customer_id
		`+where+`
		ORDER BY devices.created_at DESC, devices.id DESC
		LIMIT ?
	`, args...)
	if err != nil {
		return DevicePage{}, fmt.Errorf("list devices: %w", err)
	}
	defer rows.Close()

	var items []DeviceListItem
	for rows.Next() {
		var item DeviceListItem
		var created string
		if err := rows.Scan(&item.ID, &item.Name, &item.Brand, &item.Model, &item.CustomerName, &item.Status, &created); err != nil {
			return DevicePage{}, fmt.Errorf("scan device: %w", err)
		}
		parsed, err := time.Parse(timestampLayout, created)
		if err != nil {
			return DevicePage{}, fmt.Errorf("device time: %w", err)
		}
		item.CreatedAt = parsed
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return DevicePage{}, fmt.Errorf("list devices: %w", err)
	}
	page := DevicePage{Items: items}
	if len(items) > limit {
		page.HasMore = true
		page.Items = items[:limit]
	}
	if page.Items == nil {
		page.Items = []DeviceListItem{}
	}
	return page, nil
}

func GetDevice(ctx context.Context, database *sql.DB, id string) (Device, error) {
	return loadDevice(ctx, database, id)
}

func RegisterDevice(ctx context.Context, database *sql.DB, root string, actor User, input NewDevice) (Device, error) {
	fields := &FieldError{}
	var customerName, phone, email string
	if input.CustomerID == nil {
		customerName = requiredText(input.CustomerName, maxNameLen, "customerName", fields)
		phone = optionalText(input.CustomerPhone, maxPhoneLen, "phone", fields)
		email = optionalEmail(input.CustomerEmail, fields)
	} else if *input.CustomerID <= 0 {
		fields.addInvalid("customer")
	}
	name := requiredText(input.Name, maxNameLen, "name", fields)
	brand := requiredText(input.Brand, maxNameLen, "brand", fields)
	model := requiredText(input.Model, maxNameLen, "model", fields)
	serial := optionalText(input.Serial, maxSerialLen, "serial", fields)
	note := requiredText(input.Note, maxNoteLen, "note", fields)
	parts := make([]NewPart, len(input.Parts))
	groups := [][]int{input.DevicePhotos}
	for i, part := range input.Parts {
		parts[i] = NewPart{
			Name:   requiredText(part.Name, maxNameLen, fmt.Sprintf("parts.%d.name", i), fields),
			Note:   requiredText(part.Note, maxNoteLen, fmt.Sprintf("parts.%d.note", i), fields),
			Photos: part.Photos,
		}
		groups = append(groups, part.Photos)
	}
	claimPhotos(len(input.Files), groups, fields)
	if len(input.Files) > maxPhotos {
		fields.Photo = "limit"
	}
	if err := fields.asError(); err != nil {
		return Device{}, err
	}
	if input.CustomerID != nil {
		if _, err := loadCustomer(ctx, database, *input.CustomerID); err != nil {
			return Device{}, err
		}
	}

	deviceID, err := newID()
	if err != nil {
		return Device{}, err
	}
	photos, err := storeUploads(root, deviceID, input.Files)
	if err != nil {
		return Device{}, err
	}
	committed := false
	defer func() {
		if !committed {
			removeUploaded(root, deviceID, photoRels(photos))
		}
	}()

	tx, err := database.BeginTx(ctx, nil)
	if err != nil {
		return Device{}, fmt.Errorf("begin device: %w", err)
	}
	defer tx.Rollback()

	now := timestamp(time.Now())
	var customer Customer
	if input.CustomerID == nil {
		customer, err = insertCustomer(ctx, tx, customerName, phone, email, now)
		if err != nil {
			return Device{}, err
		}
		if err := insertAudit(ctx, tx, actor, EventCustomerCreated, customer.Name, map[string]string{
			"name": customer.Name,
		}); err != nil {
			return Device{}, err
		}
	} else {
		customer, err = loadCustomer(ctx, tx, *input.CustomerID)
		if err != nil {
			return Device{}, err
		}
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO devices (
			id, customer_id, name, brand, model, serial, note, status,
			name_fold, brand_fold, model_fold, serial_fold, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, deviceID, customer.ID, name, brand, model, serial, note, StatusRegistered,
		fold(name), fold(brand), fold(model), fold(serial), now)
	if err != nil {
		return Device{}, fmt.Errorf("insert device: %w", err)
	}

	partIDs := make([]string, len(parts))
	for i, part := range parts {
		partID, err := newID()
		if err != nil {
			return Device{}, err
		}
		partIDs[i] = partID
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO parts (id, device_id, name, note, position) VALUES (?, ?, ?, ?, ?)
		`, partID, deviceID, part.Name, part.Note, i); err != nil {
			return Device{}, fmt.Errorf("insert part: %w", err)
		}
	}
	if err := insertPhotoRows(ctx, tx, deviceID, now, photos, input.DevicePhotos, nil); err != nil {
		return Device{}, err
	}
	for i, part := range parts {
		partID := partIDs[i]
		if err := insertPhotoRows(ctx, tx, deviceID, now, photos, part.Photos, &partID); err != nil {
			return Device{}, err
		}
	}
	if err := insertAudit(ctx, tx, actor, EventDeviceRegistered, name+" "+deviceID+" "+customer.Name, map[string]string{
		"name":     name,
		"uuid":     deviceID,
		"customer": customer.Name,
	}); err != nil {
		return Device{}, err
	}
	if err := tx.Commit(); err != nil {
		return Device{}, fmt.Errorf("commit device: %w", err)
	}
	committed = true
	return loadDevice(ctx, database, deviceID)
}

func UpdateDevice(ctx context.Context, database *sql.DB, root string, actor User, id string, input DeviceChange) (Device, error) {
	current, err := loadDevice(ctx, database, id)
	if err != nil {
		return Device{}, err
	}

	fields := &FieldError{}
	name := requiredText(input.Name, maxNameLen, "name", fields)
	brand := requiredText(input.Brand, maxNameLen, "brand", fields)
	model := requiredText(input.Model, maxNameLen, "model", fields)
	serial := optionalText(input.Serial, maxSerialLen, "serial", fields)
	note := requiredText(input.Note, maxNoteLen, "note", fields)
	customerName := requiredText(input.CustomerName, maxNameLen, "customerName", fields)
	phone := optionalText(input.CustomerPhone, maxPhoneLen, "phone", fields)
	email := optionalEmail(input.CustomerEmail, fields)
	groups := [][]int{input.DevicePhotos}
	for i, part := range input.Parts {
		if strings.TrimSpace(part.Name) == "" {
			fields.addMissing(fmt.Sprintf("parts.%d.name", i))
		} else if len([]rune(strings.TrimSpace(part.Name))) > maxNameLen {
			fields.addTooLong(fmt.Sprintf("parts.%d.name", i))
		}
		if strings.TrimSpace(part.Note) == "" {
			fields.addMissing(fmt.Sprintf("parts.%d.note", i))
		} else if len([]rune(strings.TrimSpace(part.Note))) > maxNoteLen {
			fields.addTooLong(fmt.Sprintf("parts.%d.note", i))
		}
		groups = append(groups, part.Photos)
		input.Parts[i].Name = strings.TrimSpace(part.Name)
		input.Parts[i].Note = strings.TrimSpace(part.Note)
	}
	claimPhotos(len(input.Files), groups, fields)
	if err := fields.asError(); err != nil {
		return Device{}, err
	}
	input.Name, input.Brand, input.Model, input.Serial, input.Note = name, brand, model, serial, note
	input.CustomerName, input.CustomerPhone, input.CustomerEmail = customerName, phone, email
	photoCount := len(input.Files) + len(input.DeviceKeep)
	for _, part := range input.Parts {
		photoCount += len(part.Keep)
	}
	if photoCount > maxPhotos {
		return Device{}, &FieldError{Photo: "limit"}
	}

	if deviceUnchanged(current, input) {
		return current, nil
	}

	var photos []storedPhoto
	if len(input.Files) > 0 {
		photos, err = storeUploads(root, id, input.Files)
		if err != nil {
			return Device{}, err
		}
	}
	committed := false
	defer func() {
		if !committed {
			removeUploaded(root, "", photoRels(photos))
		}
	}()

	tx, err := database.BeginTx(ctx, nil)
	if err != nil {
		return Device{}, fmt.Errorf("begin device update: %w", err)
	}
	defer tx.Rollback()

	current, err = loadDevice(ctx, tx, id)
	if err != nil {
		return Device{}, err
	}
	if _, err := writeCustomer(ctx, tx, actor, current.Customer, customerName, phone, email); err != nil {
		return Device{}, err
	}
	removed, err := applyDevice(ctx, tx, actor, root, current, input, photos)
	if err != nil {
		return Device{}, err
	}
	if err := tx.Commit(); err != nil {
		return Device{}, fmt.Errorf("commit device update: %w", err)
	}
	committed = true
	removePaths(removed)
	return loadDevice(ctx, database, id)
}

func OpenPhoto(ctx context.Context, database *sql.DB, root string, deviceID string, photoID string) (string, string, error) {
	var rel string
	var media string
	err := database.QueryRowContext(ctx, `
		SELECT path, media_type FROM photos WHERE id = ? AND device_id = ?
	`, photoID, deviceID).Scan(&rel, &media)
	if err == sql.ErrNoRows {
		return "", "", ErrNotFound
	}
	if err != nil {
		return "", "", fmt.Errorf("read photo: %w", err)
	}
	full, err := photoFull(root, rel)
	if err != nil {
		return "", "", err
	}
	return full, media, nil
}

func writeCustomer(ctx context.Context, tx *sql.Tx, actor User, current Customer, name string, phone string, email string) (Customer, error) {
	if current.Name == name && current.Phone == phone && current.Email == email {
		return current, nil
	}
	if _, err := tx.ExecContext(ctx, `
		UPDATE customers
		SET name = ?, phone = ?, email = ?, name_fold = ?, phone_fold = ?, email_fold = ?
		WHERE id = ?
	`, name, phone, email, fold(name), fold(phone), fold(email), current.ID); err != nil {
		return Customer{}, fmt.Errorf("update customer: %w", err)
	}
	detail := map[string]string{"name": name}
	subject := name
	if current.Name != name {
		detail["previousName"] = current.Name
		subject = current.Name + " " + name
	}
	if err := insertAudit(ctx, tx, actor, EventCustomerUpdated, subject, detail); err != nil {
		return Customer{}, err
	}
	current.Name = name
	current.Phone = phone
	current.Email = email
	return current, nil
}

func applyDevice(ctx context.Context, tx *sql.Tx, actor User, root string, current Device, input DeviceChange, photos []storedPhoto) ([]string, error) {
	kept := map[string]bool{}
	for _, id := range input.DeviceKeep {
		kept[id] = true
	}
	currentParts := map[string]Part{}
	for _, part := range current.Parts {
		currentParts[part.ID] = part
	}
	seenPart := map[string]bool{}
	for _, part := range input.Parts {
		if part.ID == "" {
			continue
		}
		if seenPart[part.ID] {
			return nil, &FieldError{Invalid: []string{"part"}}
		}
		seenPart[part.ID] = true
		if _, ok := currentParts[part.ID]; !ok {
			return nil, &FieldError{Invalid: []string{"part"}}
		}
		for _, photoID := range part.Keep {
			kept[photoID] = true
		}
	}

	photoPart := map[string]string{}
	for _, photo := range current.Photos {
		photoPart[photo.ID] = ""
	}
	for _, part := range current.Parts {
		for _, photo := range part.Photos {
			photoPart[photo.ID] = part.ID
		}
	}
	if len(input.DeviceKeep) != len(distinct(input.DeviceKeep)) {
		return nil, &FieldError{Invalid: []string{"photos"}}
	}
	for _, photoID := range input.DeviceKeep {
		partID, ok := photoPart[photoID]
		if !ok || partID != "" {
			return nil, &FieldError{Invalid: []string{"photos"}}
		}
	}
	for _, part := range input.Parts {
		if part.ID == "" {
			if len(part.Keep) > 0 {
				return nil, &FieldError{Invalid: []string{"photos"}}
			}
			continue
		}
		if len(part.Keep) != len(distinct(part.Keep)) {
			return nil, &FieldError{Invalid: []string{"photos"}}
		}
		for _, photoID := range part.Keep {
			owner, ok := photoPart[photoID]
			if !ok || owner != part.ID {
				return nil, &FieldError{Invalid: []string{"photos"}}
			}
		}
	}

	finalCount := len(photos) + len(input.DeviceKeep)
	for _, part := range input.Parts {
		finalCount += len(part.Keep)
	}
	if finalCount > maxPhotos {
		return nil, &FieldError{Photo: "limit"}
	}

	var removed []string
	for id := range photoPart {
		if kept[id] {
			continue
		}
		var rel string
		err := tx.QueryRowContext(ctx, `SELECT path FROM photos WHERE id = ? AND device_id = ?`, id, current.ID).Scan(&rel)
		if err != nil {
			return nil, fmt.Errorf("read removed photo: %w", err)
		}
		full, err := photoFull(root, rel)
		if err != nil {
			return nil, err
		}
		removed = append(removed, full)
		if _, err := tx.ExecContext(ctx, `DELETE FROM photos WHERE id = ? AND device_id = ?`, id, current.ID); err != nil {
			return nil, fmt.Errorf("delete photo: %w", err)
		}
	}
	for _, part := range current.Parts {
		if seenPart[part.ID] {
			continue
		}
		if _, err := tx.ExecContext(ctx, `DELETE FROM parts WHERE id = ? AND device_id = ?`, part.ID, current.ID); err != nil {
			return nil, fmt.Errorf("delete part: %w", err)
		}
	}

	changed := current.Name != input.Name || current.Brand != input.Brand || current.Model != input.Model ||
		current.Serial != input.Serial || current.Note != input.Note || len(removed) > 0 || len(photos) > 0 ||
		len(current.Parts) != len(input.Parts)
	if !changed {
		for i, part := range current.Parts {
			next := input.Parts[i]
			if part.ID != next.ID || part.Name != next.Name || part.Note != next.Note {
				changed = true
				break
			}
		}
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE devices
		SET name = ?, brand = ?, model = ?, serial = ?, note = ?,
			name_fold = ?, brand_fold = ?, model_fold = ?, serial_fold = ?
		WHERE id = ?
	`, input.Name, input.Brand, input.Model, input.Serial, input.Note,
		fold(input.Name), fold(input.Brand), fold(input.Model), fold(input.Serial), current.ID); err != nil {
		return nil, fmt.Errorf("update device: %w", err)
	}

	now := timestamp(time.Now())
	for i, part := range input.Parts {
		partID := part.ID
		if partID == "" {
			nextID, err := newID()
			if err != nil {
				return nil, err
			}
			partID = nextID
			changed = true
			if _, err := tx.ExecContext(ctx, `
				INSERT INTO parts (id, device_id, name, note, position) VALUES (?, ?, ?, ?, ?)
			`, partID, current.ID, part.Name, part.Note, i); err != nil {
				return nil, fmt.Errorf("insert part: %w", err)
			}
		} else if _, err := tx.ExecContext(ctx, `
			UPDATE parts SET name = ?, note = ?, position = ? WHERE id = ? AND device_id = ?
		`, part.Name, part.Note, i, partID, current.ID); err != nil {
			return nil, fmt.Errorf("update part: %w", err)
		}
		if err := insertPhotoRows(ctx, tx, current.ID, now, photos, part.Photos, &partID); err != nil {
			return nil, err
		}
	}
	if err := insertPhotoRows(ctx, tx, current.ID, now, photos, input.DevicePhotos, nil); err != nil {
		return nil, err
	}
	if changed {
		if err := insertAudit(ctx, tx, actor, EventDeviceUpdated, input.Name+" "+current.ID+" "+input.CustomerName, map[string]string{
			"name":     input.Name,
			"uuid":     current.ID,
			"customer": input.CustomerName,
		}); err != nil {
			return nil, err
		}
	}
	return removed, nil
}

func deviceUnchanged(current Device, input DeviceChange) bool {
	if current.Name != input.Name || current.Brand != input.Brand || current.Model != input.Model ||
		current.Serial != input.Serial || current.Note != input.Note {
		return false
	}
	if current.Customer.Name != input.CustomerName || current.Customer.Phone != input.CustomerPhone || current.Customer.Email != input.CustomerEmail {
		return false
	}
	if len(input.Files) > 0 || len(current.Parts) != len(input.Parts) {
		return false
	}
	if !sameIDs(photoIDs(current.Photos), input.DeviceKeep) {
		return false
	}
	for i, part := range current.Parts {
		next := input.Parts[i]
		if part.ID == "" || part.ID != next.ID || part.Name != next.Name || part.Note != next.Note || len(next.Photos) > 0 {
			return false
		}
		if !sameIDs(photoIDs(part.Photos), next.Keep) {
			return false
		}
	}
	return true
}

func insertCustomer(ctx context.Context, tx *sql.Tx, name string, phone string, email string, now string) (Customer, error) {
	result, err := tx.ExecContext(ctx, `
		INSERT INTO customers (name, phone, email, name_fold, phone_fold, email_fold, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, name, phone, email, fold(name), fold(phone), fold(email), now)
	if err != nil {
		return Customer{}, fmt.Errorf("insert customer: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return Customer{}, fmt.Errorf("customer id: %w", err)
	}
	created, err := time.Parse(timestampLayout, now)
	if err != nil {
		return Customer{}, err
	}
	return Customer{ID: id, Name: name, Phone: phone, Email: email, CreatedAt: created}, nil
}

func loadCustomer(ctx context.Context, q queryer, id int64) (Customer, error) {
	row := q.QueryRowContext(ctx, `
		SELECT id, name, phone, email, created_at FROM customers WHERE id = ?
	`, id)
	customer, err := scanCustomer(row)
	if err == sql.ErrNoRows {
		return Customer{}, ErrNotFound
	}
	if err != nil {
		return Customer{}, fmt.Errorf("read customer: %w", err)
	}
	return customer, nil
}

type scanner interface {
	Scan(dest ...any) error
}

func scanCustomer(row scanner) (Customer, error) {
	var customer Customer
	var created string
	if err := row.Scan(&customer.ID, &customer.Name, &customer.Phone, &customer.Email, &created); err != nil {
		return Customer{}, err
	}
	parsed, err := time.Parse(timestampLayout, created)
	if err != nil {
		return Customer{}, fmt.Errorf("customer time: %w", err)
	}
	customer.CreatedAt = parsed
	return customer, nil
}

func loadDevice(ctx context.Context, q queryer, id string) (Device, error) {
	var device Device
	var created string
	var customerCreated string
	err := q.QueryRowContext(ctx, `
		SELECT devices.id, devices.name, devices.brand, devices.model, devices.serial, devices.note,
			devices.status, devices.created_at, customers.id, customers.name, customers.phone, customers.email, customers.created_at
		FROM devices
		JOIN customers ON customers.id = devices.customer_id
		WHERE devices.id = ?
	`, id).Scan(
		&device.ID, &device.Name, &device.Brand, &device.Model, &device.Serial, &device.Note,
		&device.Status, &created, &device.Customer.ID, &device.Customer.Name, &device.Customer.Phone,
		&device.Customer.Email, &customerCreated,
	)
	if err == sql.ErrNoRows {
		return Device{}, ErrNotFound
	}
	if err != nil {
		return Device{}, fmt.Errorf("read device: %w", err)
	}
	device.CreatedAt, err = time.Parse(timestampLayout, created)
	if err != nil {
		return Device{}, fmt.Errorf("device time: %w", err)
	}
	device.Customer.CreatedAt, err = time.Parse(timestampLayout, customerCreated)
	if err != nil {
		return Device{}, fmt.Errorf("customer time: %w", err)
	}

	device.Parts = []Part{}
	device.Photos = []Photo{}
	partRows, err := q.QueryContext(ctx, `
		SELECT id, name, note FROM parts WHERE device_id = ? ORDER BY position, id
	`, id)
	if err != nil {
		return Device{}, fmt.Errorf("list parts: %w", err)
	}
	defer partRows.Close()
	parts := map[string]*Part{}
	for partRows.Next() {
		var part Part
		if err := partRows.Scan(&part.ID, &part.Name, &part.Note); err != nil {
			return Device{}, fmt.Errorf("scan part: %w", err)
		}
		part.Photos = []Photo{}
		device.Parts = append(device.Parts, part)
	}
	if err := partRows.Err(); err != nil {
		return Device{}, fmt.Errorf("list parts: %w", err)
	}
	for i := range device.Parts {
		parts[device.Parts[i].ID] = &device.Parts[i]
	}

	photoRows, err := q.QueryContext(ctx, `
		SELECT id, part_id, path, size, media_type
		FROM photos
		WHERE device_id = ?
		ORDER BY created_at, id
	`, id)
	if err != nil {
		return Device{}, fmt.Errorf("list photos: %w", err)
	}
	defer photoRows.Close()
	for photoRows.Next() {
		var photo Photo
		var partID sql.NullString
		if err := photoRows.Scan(&photo.ID, &partID, &photo.Path, &photo.Size, &photo.MediaType); err != nil {
			return Device{}, fmt.Errorf("scan photo: %w", err)
		}
		if partID.Valid {
			part := parts[partID.String]
			if part == nil {
				return Device{}, fmt.Errorf("photo part %s", partID.String)
			}
			part.Photos = append(part.Photos, photo)
			continue
		}
		device.Photos = append(device.Photos, photo)
	}
	if err := photoRows.Err(); err != nil {
		return Device{}, fmt.Errorf("list photos: %w", err)
	}
	return device, nil
}

func claimPhotos(count int, groups [][]int, fields *FieldError) {
	seen := map[int]bool{}
	claimed := 0
	for _, group := range groups {
		for _, index := range group {
			if index < 0 || index >= count || seen[index] {
				fields.addInvalid("photos")
				return
			}
			seen[index] = true
			claimed++
		}
	}
	if claimed != count {
		fields.addInvalid("photos")
	}
}

func insertPhotoRows(ctx context.Context, tx *sql.Tx, deviceID string, now string, photos []storedPhoto, indexes []int, partID *string) error {
	for _, index := range indexes {
		photo := photos[index]
		var part any
		if partID != nil {
			part = *partID
		}
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO photos (id, device_id, part_id, path, size, media_type, created_at)
			VALUES (?, ?, ?, ?, ?, ?, ?)
		`, photo.id, deviceID, part, photo.rel, photo.size, photo.media, now); err != nil {
			return fmt.Errorf("insert photo: %w", err)
		}
	}
	return nil
}

func storeUploads(root string, deviceID string, files []Upload) ([]storedPhoto, error) {
	var done []storedPhoto
	for _, file := range files {
		reader, err := file.Open()
		if err != nil {
			removeUploaded(root, "", photoRels(done))
			return nil, fmt.Errorf("open upload: %w", err)
		}
		photo, err := writePhoto(root, deviceID, reader)
		reader.Close()
		if err != nil {
			removeUploaded(root, "", photoRels(done))
			return nil, err
		}
		done = append(done, photo)
	}
	return done, nil
}

func photoRels(photos []storedPhoto) []string {
	rels := make([]string, 0, len(photos))
	for _, photo := range photos {
		rels = append(rels, photo.rel)
	}
	return rels
}

func removeUploaded(root string, deviceID string, rels []string) {
	fulls := make([]string, 0, len(rels))
	for _, rel := range rels {
		full, err := photoFull(root, rel)
		if err == nil {
			fulls = append(fulls, full)
		}
	}
	removePaths(fulls)
	if deviceID == "" {
		return
	}
	dir, err := photoFull(root, "devices/"+deviceID)
	if err == nil {
		_ = os.Remove(dir)
	}
}

func photoIDs(photos []Photo) []string {
	ids := make([]string, len(photos))
	for i, photo := range photos {
		ids[i] = photo.ID
	}
	return ids
}

func sameIDs(left []string, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	seen := map[string]int{}
	for _, id := range left {
		seen[id]++
	}
	for _, id := range right {
		seen[id]--
		if seen[id] < 0 {
			return false
		}
	}
	return true
}

func distinct(ids []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, id := range ids {
		if seen[id] {
			continue
		}
		seen[id] = true
		out = append(out, id)
	}
	return out
}
