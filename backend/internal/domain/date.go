package domain

import (
	"database/sql/driver"
	"fmt"
	"time"
)

// dateLayout is the wire format used for all date fields in the API:
// "YYYY-MM-DD".
const dateLayout = "2006-01-02"

// Date wraps time.Time so it marshals/unmarshals as a bare "YYYY-MM-DD"
// string over JSON, while still scanning/valuing correctly against
// Postgres `date` columns via database/sql (and pgx's sql.Scanner fallback).
type Date struct {
	time.Time
}

// NewDate builds a Date from a time.Time, truncating to the date portion.
func NewDate(t time.Time) Date {
	return Date{Time: time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)}
}

// ParseDate parses a "YYYY-MM-DD" string into a Date.
func ParseDate(s string) (Date, error) {
	t, err := time.Parse(dateLayout, s)
	if err != nil {
		return Date{}, fmt.Errorf("invalid date %q: expected YYYY-MM-DD", s)
	}
	return Date{Time: t}, nil
}

// String formats the date as "YYYY-MM-DD".
func (d Date) String() string {
	return d.Time.Format(dateLayout)
}

// MarshalJSON emits the date as a bare "YYYY-MM-DD" JSON string.
func (d Date) MarshalJSON() ([]byte, error) {
	return []byte(`"` + d.Time.Format(dateLayout) + `"`), nil
}

// UnmarshalJSON parses a "YYYY-MM-DD" JSON string.
func (d *Date) UnmarshalJSON(data []byte) error {
	s := string(data)
	if s == "null" {
		d.Time = time.Time{}
		return nil
	}
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		s = s[1 : len(s)-1]
	}
	t, err := time.Parse(dateLayout, s)
	if err != nil {
		return fmt.Errorf("invalid date %q: expected YYYY-MM-DD", s)
	}
	d.Time = t
	return nil
}

// Scan implements sql.Scanner so Date can be read directly from a
// Postgres `date` column.
func (d *Date) Scan(src any) error {
	if src == nil {
		d.Time = time.Time{}
		return nil
	}
	switch v := src.(type) {
	case time.Time:
		d.Time = v
		return nil
	case string:
		t, err := time.Parse(dateLayout, v)
		if err != nil {
			return err
		}
		d.Time = t
		return nil
	case []byte:
		t, err := time.Parse(dateLayout, string(v))
		if err != nil {
			return err
		}
		d.Time = t
		return nil
	default:
		return fmt.Errorf("unsupported Scan type for Date: %T", src)
	}
}

// Value implements driver.Valuer so Date can be written directly into a
// Postgres `date` column.
func (d Date) Value() (driver.Value, error) {
	return d.Time, nil
}
