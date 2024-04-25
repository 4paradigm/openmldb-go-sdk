package openmldb

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"time"
)

var (
	_ sql.Scanner = (*NullDate)(nil)
	_ driver.Valuer = NullDate{}
)

type NullDate struct {
	Time  time.Time
	Valid bool // Valid is true if Time is not NULL
}

// Scan implements sql.Scanner for NullDate
func (dt *NullDate) Scan(src any) error {
	switch val := src.(type) {
	case string:
		dval, err := time.Parse(time.DateOnly, val)
		if err != nil {
			dt.Valid = false
			return err
		} else {
			dt.Time = dval
			dt.Valid = true
			return nil
		}
	case NullDate:
		*dt = val
		return nil
	default:
		return errors.New("scan NullDate from unsupported type")
	}

}

// Value implements driver.Valuer for NullDate
func (dt NullDate) Value() (driver.Value, error) {
	if !dt.Valid {
		return nil, nil
	}
	return dt.Time, nil

}
