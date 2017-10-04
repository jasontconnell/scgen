package data

import (
	"time"
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

type Item struct {
	Parent   *Item
	Children []*Item

	ID                string
	Name              string
	CleanName         string
	TemplateID        string
	ParentID          string
	MasterID          string
	Path              string
	Created           time.Time
	Updated           time.Time
	FieldType         string
	BaseTemplates     string
	BaseTemplateItems []*Item
}

func (i Item) String() string {
	return "ID: " + i.ID + " Name:" + i.Name + " Path: " + i.Path
}

type Template struct {
	Item                 *Item
	BaseTemplates        []*Template
	Fields               []Field
	Namespace            string
	ReferencedNamespaces []string
}

type Field struct {
	Item      *Item
	Name      string
	CleanName string
	TypeName  string
	CodeType  string
	Suffix    string
}

type DeserializedItem struct {
	ID         string
	Name       string
	TemplateID string
	ParentID   string
	MasterID   string
	Fields     []DeserializedField
}

type DeserializedField struct {
	ID       string
	Name     string
	Value    string
	Version  int64
	Language string
	Source   string
}

type UpdateType string

const (
	Insert UpdateType = "insert"
	Update UpdateType = "update"
	Delete UpdateType = "delete"
)

type UpdateItem struct {
	ID         string
	Name       string
	TemplateID string
	ParentID   string
	MasterID   string
	UpdateType UpdateType
}

type UpdateField struct {
	ItemID     string
	FieldID    string
	Value      string
	Source     string
	Version    int64
	Language   string
	UpdateType UpdateType
}
