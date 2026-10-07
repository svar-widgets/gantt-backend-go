package data

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"gantt-backend-go/common"
)

type Segment struct {
	Text     string        `json:"text"`
	Start    *common.JDate `json:"start"`
	End      *common.JDate `json:"end"`
	Duration int           `json:"duration"`
}

type Segments []Segment

func (s Segments) Value() (driver.Value, error) {
	if len(s) == 0 {
		return nil, nil
	}
	bytes, err := json.Marshal(s)
	return string(bytes), err
}

func (s *Segments) Scan(value any) error {
	*s = nil
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		str, ok := value.(string)
		if !ok {
			return fmt.Errorf("failed to unmarshal segments value: %v", value)
		}
		bytes = []byte(str)
	}
	return json.Unmarshal(bytes, s)
}

func (Segments) GormDataType() string {
	return "text"
}
