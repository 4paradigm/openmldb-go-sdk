package openmldb

// TODO(someone): support go < 1.22

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"time"
)

var (
	_ sql.Scanner   = (*NullDate)(nil)
	_ driver.Valuer = NullDate{}
)

// Null represents a value that may be null.
//
// declare type embedded sql.Null so we still able to
// utilize sql.Scanner and driver.Valuer in go standard,
// and customize marshal logic for api requests
type Null[T any] struct {
	sql.Null[T]
}

// NullDate represents nullable SQL date in go
//
// embedded sql.Null[time.Time] so it by default
// implements sql.Scanner and driver.Valuer, but
// distinct timestamp representation in sdk.
type NullDate struct {
	sql.Null[time.Time]
}

// MarshalJSON implements json.Marshaler
func (src NullDate) MarshalJSON() ([]byte, error) {
	if !src.Valid {
		return json.Marshal(nil)
	}
	return json.Marshal(src.V.Format(time.DateOnly))
}

// MarshalJSON implements json.Marshaler for Null[T]
func (src Null[T]) MarshalJSON() ([]byte, error) {
	if !src.Valid {
		return json.Marshal(nil)
	}

	var v any = src.V
	switch val := v.(type) {
	case time.Time:
		// timestamp, marshal to int64 unix epoch time in millisecond
		return json.Marshal(val.UnixMilli())
	default:
		return json.Marshal(src.V)
	}
}
