package model

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
