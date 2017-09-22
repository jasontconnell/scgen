package data

import (
    "time"
)
type Item struct {
    Parent *Item
    Children []*Item

    ID string
    Name string
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
}

type Field struct {
    Name string
    Type string
}

