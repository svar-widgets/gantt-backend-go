package data

import (
	"encoding/json"
	"gantt-backend-go/common"
)

type Task struct {
	ID                    int           `json:"id"`
	Text                  string        `json:"text"`
	Details               string        `json:"details"`
	Start                 *common.JDate `json:"start"`
	End                   *common.JDate `json:"end"`
	Duration              int           `json:"duration"`
	Progress              int           `json:"progress"`
	Parent                int           `json:"parent"`
	Type                  string        `json:"type"`
	Lazy                  bool          `json:"lazy"`
	Open                  bool          `json:"open"`
	Constraint            *Constraint   `json:"constraint"`
	Deadline              *common.JDate `json:"deadline"`
	BaseStart             *common.JDate `json:"base_start"`
	BaseEnd               *common.JDate `json:"base_end"`
	BaseDuration          int           `json:"base_duration"`
	Calendar              *common.ID    `json:"calendar"`
	SkipResourceCalendars bool          `json:"skipResourceCalendars"`
	Unscheduled           bool          `json:"unscheduled"`
	Manual                bool          `json:"manual"`
	Inactive              bool          `json:"inactive"`
	Rollup                bool          `json:"rollup"`
	Segments              Segments      `json:"segments"`
	Index                 int           `json:"-"`
}

type Link struct {
	ID     int     `json:"id"`
	Source int     `json:"source"`
	Target int     `json:"target"`
	Type   string  `json:"type"`
	Lag    float64 `json:"lag"`
}

type Resource struct {
	ID       int        `json:"id"`
	Name     string     `json:"name"`
	Avatar   string     `json:"avatar"`
	Color    string     `json:"color"`
	Role     string     `json:"role"`
	Calendar *common.ID `json:"calendar"`
	Parent   int        `json:"parent"`
}

type Assignment struct {
	ID       int     `json:"id"`
	Task     int     `json:"task"`
	Resource int     `json:"resource"`
	Units    float64 `json:"units"`
}

type BatchItem struct {
	Url    string          `json:"url"`
	Method string          `json:"method"`
	Data   json.RawMessage `json:"data"`
}
