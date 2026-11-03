package store

import (
	"net/mail"
	"strings"
	"time"
)

type FieldError struct {
	Missing []string
	Invalid []string
	TooLong []string
	Photo   string
}

func (e *FieldError) Error() string {
	return "validation"
}

func (e *FieldError) addMissing(field string) {
	e.Missing = append(e.Missing, field)
}

func (e *FieldError) addInvalid(field string) {
	e.Invalid = append(e.Invalid, field)
}

func (e *FieldError) addTooLong(field string) {
	e.TooLong = append(e.TooLong, field)
}

func (e *FieldError) empty() bool {
	return len(e.Missing) == 0 && len(e.Invalid) == 0 && len(e.TooLong) == 0 && e.Photo == ""
}

func (e *FieldError) asError() error {
	if e.empty() {
		return nil
	}
	return e
}

func timestamp(t time.Time) string {
	return t.UTC().Format(timestampLayout)
}

func FormatTime(t time.Time) string {
	return timestamp(t)
}

func pageLimit(n int) (int, error) {
	if n == 0 {
		return defaultPage, nil
	}
	if n < 1 || n > maxPage {
		return 0, ErrInvalidLimit
	}
	return n, nil
}

func fold(value string) string {
	value = strings.TrimSpace(value)
	value = strings.NewReplacer(
		"İ", "i", "I", "i", "ı", "i",
		"Ş", "s", "ş", "s",
		"Ğ", "g", "ğ", "g",
		"Ü", "u", "ü", "u",
		"Ö", "o", "ö", "o",
		"Ç", "c", "ç", "c",
	).Replace(value)
	return strings.ToLower(value)
}

func likePattern(value string) string {
	var b strings.Builder
	b.WriteByte('%')
	for _, r := range fold(value) {
		if r == '%' || r == '_' || r == '\\' {
			b.WriteByte('\\')
		}
		b.WriteRune(r)
	}
	b.WriteByte('%')
	return b.String()
}

func requiredText(value string, max int, field string, fields *FieldError) string {
	value = strings.TrimSpace(value)
	if value == "" {
		fields.addMissing(field)
		return ""
	}
	if len([]rune(value)) > max {
		fields.addTooLong(field)
	}
	return value
}

func optionalText(value string, max int, field string, fields *FieldError) string {
	value = strings.TrimSpace(value)
	if len([]rune(value)) > max {
		fields.addTooLong(field)
	}
	return value
}

func optionalEmail(value string, fields *FieldError) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	if len([]rune(value)) > maxEmailLen {
		fields.addTooLong("email")
		return value
	}
	parsed, err := mail.ParseAddress(value)
	if err != nil || parsed.Address != value || !strings.Contains(value, ".") {
		fields.addInvalid("email")
	}
	return value
}
