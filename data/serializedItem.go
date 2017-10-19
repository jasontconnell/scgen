package data

import (
	"fmt"
)

type SerializedItem struct {
	Item   *Item
	Fields []SerializedField
}

type SerializedField struct {
	FieldValueID string
	FieldID      string
	Name         string
	Value        string
	Version      int64
	Language     string
	Source       string
}

func (item SerializedItem) String() string {
	fields := ""
	for _, f := range item.Fields {
		fields += fmt.Sprintf("\nFields\n    Name: %v  Value length: %v", f.Name, len(f.Value))
	}
	s := fmt.Sprintf("Name: %v  ID: %v ", item.Item.Name, item.Item.ID)

	return s + fields
}