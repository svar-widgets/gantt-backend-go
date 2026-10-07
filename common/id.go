package common

import (
	"database/sql/driver"
	"encoding/json"
)

// ID keeps a client-set reference as it came, a string or an integer, so it is
// sent back with the same type. The type survives storage through SQLite's
// dynamic typing; a strictly typed database would return it as a string.
type ID struct {
	v any
}

func (id *ID) UnmarshalJSON(data []byte) error {
	switch {
	case string(data) == "null":
		id.v = nil
	case data[0] == QuotesByte:
		var s string
		if err := json.Unmarshal(data, &s); err != nil {
			return err
		}
		id.v = s
	default:
		var n int64
		if err := json.Unmarshal(data, &n); err != nil {
			return err
		}
		id.v = n
	}
	return nil
}

func (id ID) MarshalJSON() ([]byte, error) { return json.Marshal(id.v) }

func (id ID) Value() (driver.Value, error) { return id.v, nil }

func (id *ID) Scan(src any) error {
	switch v := src.(type) {
	case []byte:
		id.v = string(v)
	default:
		id.v = v
	}
	return nil
}

func (ID) GormDataType() string { return "blob" }
