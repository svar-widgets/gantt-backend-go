package common

import (
	"database/sql/driver"
	"fmt"
	"strings"
	"time"
)

type JDate time.Time

const layout = "2006-01-02 15:04:05"

var nilTime = (time.Time{}).UnixNano()

func (d *JDate) UnmarshalJSON(b []byte) (err error) {
	s := strings.Trim(string(b), "\"")
	if s == "null" {
		*d = JDate(time.Time{})
		return
	}
	t, err := time.Parse(layout, s)
	if err != nil {
		return err
	}
	*d = JDate(t)

	return
}

func (d *JDate) MarshalJSON() ([]byte, error) {
	if d == nil {
		return []byte("null"), nil
	}
	if time.Time(*d).UnixNano() == nilTime {
		return []byte("null"), nil
	}
	return []byte(fmt.Sprintf("\"%s\"", time.Time(*d).Format(layout))), nil
}

func (d JDate) Value() (driver.Value, error) {
	return driver.Value(time.Time(d)), nil
}

func (d *JDate) Scan(src interface{}) error {
	t, ok := src.(time.Time)
	if !ok {
		return fmt.Errorf("cannot parse value to JDate")
	}

	*d = JDate(t)
	return nil
}
