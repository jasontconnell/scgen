package conf

import (
    "conf"
    "strings"
)

type Configuration struct {
    TemplateID string `json:"template"`
    TemplateFolderID string `json:"templateFolder"`
    TemplateSectionID string `json:"templateSection"`
    TemplateFieldID string `json:"templateField"`
    FieldTypes []FieldType `json:"fieldTypes"`

    FieldTypeMap map[string]string

    ConnectionString string `json:"connectionString"`
}

type FieldType struct {
    TypeName string `json:"typeName"`
    Type string `json:"type"`
}

func LoadConfig(file string) Configuration {
    config := Configuration{}
    conf.LoadConfig(file, &config)

    config.FieldTypeMap = make(map[string]string)
    for _, ft := range config.FieldTypes {
        key := strings.ToLower(ft.TypeName)
        config.FieldTypeMap[key] = ft.Type
    }

    return config
}
