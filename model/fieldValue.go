package model

type FieldValue struct {
	FieldValueID string
	ItemID       string
	FieldName    string
	FieldID      string
	Path         string
	Value        string
	Source       string
	Language     string
	Version      int64
}
