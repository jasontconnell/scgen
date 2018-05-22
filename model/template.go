package model

import (
	"github.com/google/uuid"
	"github.com/jasontconnell/sitecore/data"
)

// unbind from "sitecore" package
type Template struct {
	ID     uuid.UUID
	Name   string
	Path   string
	Parent data.ItemNode

	CleanName            string
	Namespace            string
	AlternateNamespace   string
	ReferencedNamespaces []string
	Generate             bool

	BaseTemplates []*Template
	Fields        []*Field
	Ignore        bool
	Include       bool
}

type Field struct {
	ID   uuid.UUID
	Name string

	FieldType string
	CleanName string
	CodeType  string
	Suffix    string
}
