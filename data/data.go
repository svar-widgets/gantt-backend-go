package data

import (
	"gantt-backend-go/common"
)

type Task struct {
	ID       int           `json:"id"`
	Text     string        `json:"text"`
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
