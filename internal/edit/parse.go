package edit

import (
	"encoding/json"
	"fmt"
	"time"
)

// mirrors part of wikipedias recentchange JSON
type rawEdit struct {
	Type   string `json:"type"`
	Title  string `json:"title"`
	User   string `json:"user"`
	Bot    bool   `json:"bot"`
	Wiki   string `json:"wiki"`
	Length *struct {
		Old int `json:"old"`
		New int `json:"new"`
	} `json:"length"`
	Timestamp int64 `json:"timestamp"`
}

func ParseEdit(data []byte) (Edit, error) {
	var raw rawEdit

	if err := json.Unmarshal(data, &raw); err != nil {
		return Edit{}, fmt.Errorf("parse edit: %w", err)
	}

	delta := 0
	if raw.Length != nil {
		delta = raw.Length.New - raw.Length.Old
	}

	// parse from json into our Edit struct
	return Edit{
		Wiki:      raw.Wiki,
		Title:     raw.Title,
		User:      raw.User,
		Type:      raw.Type,
		Bot:       raw.Bot,
		ByteDelta: delta,
		Timestamp: time.Unix(raw.Timestamp, 0).UTC(),
	}, nil
}
