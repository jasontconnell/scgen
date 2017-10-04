package data

import (
    "time"
)

type SerializedItem struct {
    Item *Item
    Fields []SerializedField
}

type SerializedField struct {
    ID string
    Name string
    Value string
}

type FieldValue struct {
    ID string
    ItemName string
    FieldName string
    TemplateID string
    ParentID string
    FieldID string
    Path string
    Value string
    Created time.Time
    Updated time.Time
}

type Item struct {
    Parent *Item
    Children []*Item

    ID string
    Name string
    CleanName string
    TemplateID string
    ParentID string
    Path string
    Created time.Time
    Updated time.Time
    FieldType string
    BaseTemplates string
    BaseTemplateItems []*Item
}

func (i Item) String() string {
    return "ID: " + i.ID + " Name:" + i.Name + " Path: " + i.Path
}

type Template struct {
    Item *Item
    BaseTemplates []*Template
    Fields []Field
    Namespace string
    ReferencedNamespaces []string
}

type Field struct {
    Item *Item
    Name string
    CleanName string
    TypeName string
    CodeType string
    Suffix string
}

