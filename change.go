package goapollo

import "encoding/json"

const (
	ChangeEventAddType ChangeEventType = iota
	ChangeEventModifyType
	ChangeEventDeleteType
)

type ChangeEventType int

type ChangeEvent struct {
	Namespace      string
	Configurations map[string]*ChangeItem
}

type ChangeItem struct {
	OldValue   string
	NewValue   string
	ChangeType ChangeEventType
}

func NewChangeEvent(namespace string) *ChangeEvent {
	return &ChangeEvent{Namespace: namespace, Configurations: make(map[string]*ChangeItem)}
}

func (c *ChangeEvent) Store(key string, oldValue, newValue string) {
	item := &ChangeItem{OldValue: oldValue, NewValue: newValue}
	if oldValue == "" {
		item.ChangeType = ChangeEventAddType
	} else if newValue == "" {
		item.ChangeType = ChangeEventDeleteType
	} else if oldValue != newValue {
		item.ChangeType = ChangeEventModifyType
	} else if oldValue == newValue {
		return
	}
	c.Configurations[key] = item
}

func (c *ChangeEvent) String() string {
	b, _ := json.Marshal(c)
	return string(b)
}
