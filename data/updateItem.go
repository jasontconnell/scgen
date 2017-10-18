package data

type UpdateType string

const (
	Insert UpdateType = "insert"
	Update UpdateType = "update"
	Delete UpdateType = "delete"
	Ignore UpdateType = "ignore"
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
