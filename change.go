package goapollo

import "encoding/json"

// ChangeType for a key
type ChangeType int

const (
	// ADD a new value
	EventAdd ChangeType = 0
	// MODIFY a old value
	EventModify ChangeType = 1
	// DELETE ...
	EventDelete ChangeType = 2
)

func (c ChangeType) String() string {
	switch c {
	case EventAdd:
		return "ADD"
	case EventModify:
		return "MODIFY"
	case EventDelete:
		return "DELETE"
	}

	return "UNKNOW"
}

// ChangeEvent change event
type ChangeEvent struct {
	Namespace string
	Changes   map[string]*Change
}

func (c *ChangeEvent) String() string {
	if c == nil {
		return ""
	}
	body, err := json.Marshal(c)
	if err != nil {
		return ""
	}
	return string(body)
}

// Change represent a single key change
type Change struct {
	OldValue   string
	NewValue   string
	ChangeType ChangeType
}
