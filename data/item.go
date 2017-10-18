package data

import (
	"time"
)

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
