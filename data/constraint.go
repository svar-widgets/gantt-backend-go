package data

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"gantt-backend-go/common"
)

type Constraint struct {
	Type string        `json:"type"`
	Date *common.JDate `json:"date"`
}

var allowedTypes = map[string]struct{}{"snet": {}, "snlt": {}, "fnet": {}, "fnlt": {}, "mso": {}, "mfo": {}}

func (c Constraint) Value() (driver.Value, error) {
	if err := c.validate(); err != nil {
		return nil, err
	}
	return json.Marshal(c)
}

func (c *Constraint) Scan(value any) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		str, ok := value.(string)
		if !ok {
			return fmt.Errorf("failed to unmarshal constraint value: %v", value)
		}
		bytes = []byte(str)
	}
	return json.Unmarshal(bytes, c)
}

func (c Constraint) validate() error {
	if c.Date.IsZero() {
		return fmt.Errorf("invalid constraint: date is missing")
	}
	if _, ok := allowedTypes[c.Type]; !ok {
		return fmt.Errorf("invalid constraint: type %q is not allowed", c.Type)
	}
	return nil
}
