package data

type Template struct {
	Item                 *Item
	BaseTemplates        []*Template
	Fields               []Field
	Namespace            string
	AlternateNamespace   string
	ReferencedNamespaces []string
	Generate             bool
}

type Field struct {
	Item      *Item
	Name      string
	CleanName string
	TypeName  string
	CodeType  string
	Suffix    string
}
