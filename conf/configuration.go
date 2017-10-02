package conf

import (
    "conf"
    "strings"
)

type FileMode int
const (
    One FileMode = iota
    Many
)

type Configuration struct {
    TemplateID string `json:"template"`
    TemplateFolderID string `json:"templateFolder"`
    TemplateSectionID string `json:"templateSection"`
    TemplateFieldID string `json:"templateField"`
    FieldTypes []FieldType `json:"fieldTypes"`

    FieldTypeMap map[string]FieldType
    DefaultFieldType string `json:"defaultFieldType"`

    CodeTemplate string `json:"codeTemplate"`
    CodeFileExtension string `json:"codeFileExtension"`

    ConnectionString string `json:"connectionString"`

    // parameters

    BasePath string `json:"basePath"`
    BaseNamespace string `json:"baseNamespace"`
    FileModeString string `json:"filemode"`
    OutputPath string `json:"outputPath"`

    // not in config file
    FileMode FileMode
}

type FieldType struct {
    TypeName string `json:"typeName"`
    CodeType string `json:"codeType"`
    Suffix string `json:"suffix"`
}

func LoadConfig(file string) Configuration {
    config := Configuration{}
    conf.LoadConfig(file, &config)

    config.FieldTypeMap = make(map[string]FieldType)
    for _, ft := range config.FieldTypes {
        key := strings.ToLower(ft.TypeName)
        config.FieldTypeMap[key] = ft
    }

    return config
}
