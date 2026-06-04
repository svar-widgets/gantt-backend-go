package data

import (
	"encoding/json"
	"gantt-backend-go/common"
)

type Task struct {
	ID       int           `json:"id"`
	Text     string        `json:"text"`
	Details  string        `json:"details"`
	Start    *common.JDate `json:"start"`
	End      *common.JDate `json:"end"`
	Duration int           `json:"duration"`
	Progress int           `json:"progress"`
	Parent   int           `json:"parent"`
	Type     string        `json:"type"`
	Lazy     bool          `json:"lazy"`
	Open     bool          `json:"open"`
	Index    int           `json:"-"`
}

type Link struct {
	ID     int    `json:"id"`
	Source int    `json:"source"`
	Target int    `json:"target"`
	Type   string `json:"type"`
}

type Resource struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Avatar   string `json:"avatar"`
	Color    string `json:"color"`
	Role     string `json:"role"`
	Calendar string `json:"calendar"`
	Parent   int    `json:"parent"`
}

type Assignment struct {
	ID       int `json:"id"`
	Task     int `json:"task"`
	Resource int `json:"resource"`
	Units    int `json:"units"`
}

type BatchItem struct {
	Url    string          `json:"url"`
	Method string          `json:"method"`
	Data   json.RawMessage `json:"data"`
}
