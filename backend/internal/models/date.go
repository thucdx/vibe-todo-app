package models

import (
	"database/sql/driver"
	"fmt"
	"time"
)

// Date is a date-only (YYYY-MM-DD) value that round-trips cleanly through
// PostgreSQL's date column and JSON encoding. It stores a time.Time internally
// but marshals as a plain "2006-01-02" string rather than RFC 3339.
type Date time.Time

const dateLayout = "2006-01-02"

// MarshalJSON implements json.Marshaler — emits "YYYY-MM-DD".
func (d Date) MarshalJSON() ([]byte, error) {
	return []byte(`"` + time.Time(d).UTC().Format(dateLayout) + `"`), nil
}

// UnmarshalJSON implements json.Unmarshaler — accepts "YYYY-MM-DD".
func (d *Date) UnmarshalJSON(b []byte) error {
	s := string(b)
	if s == "null" {
		return nil
	}
	if len(s) < 2 {
		return fmt.Errorf("models.Date: short value %s", s)
	}
	t, err := time.ParseInLocation(dateLayout, s[1:len(s)-1], time.UTC)
	if err != nil {
		return fmt.Errorf("models.Date: %w", err)
	}
	*d = Date(t)
	return nil
}

// Value implements driver.Valuer so sqlx can write a Date into PostgreSQL.
func (d Date) Value() (driver.Value, error) { return time.Time(d), nil }

// Scan implements sql.Scanner so sqlx can read a PostgreSQL date column into Date.
func (d *Date) Scan(src any) error {
	if t, ok := src.(time.Time); ok {
		*d = Date(t)
		return nil
	}
	return fmt.Errorf("models.Date: cannot scan %T", src)
}
