package edit

import "time"

// Our Edit schema
type Edit struct {
	Wiki      string
	Title     string
	User      string
	Type      string
	Bot       bool // is it automated
	ByteDelta int  // length.new - length.old
	Timestamp time.Time
}
