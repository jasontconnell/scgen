package conf

import (
    "conf"
    "strings"
    "reflecthelper"
    "path/filepath"
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

    SerializationIgnoredFields []string `json:"serializationIgnoredFields"`
    SerializationPath string `json:"serializationPath"`
    SerializationExtension string `json:"serializationExtension"`

    ConnectionString string `json:"connectionString"`

    Serialize bool `json:"serialize"`
    Generate bool `json:"generate"`
    Deserialize bool `json:"deserialize"`

    BasePath string `json:"basePath"`
    BasePaths []string `json:"basePaths"`
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

    if len(config.FieldTypes) > 0 {
        config.FieldTypeMap = make(map[string]FieldType)
        for _, ft := range config.FieldTypes {
            key := strings.ToLower(ft.TypeName)
            config.FieldTypeMap[key] = ft
        }
    }

    if len(config.BasePaths) == 0 && len(config.BasePath) > 0 {
        config.BasePaths = append(config.BasePaths, config.BasePath)
    }

    return config
}

func LoadConfigs(workingDir, filecsv string) Configuration {
    cfgs := []Configuration{}
    configFiles := strings.Split(filecsv, ",")
    for _, file := range configFiles {
        cfg := LoadConfig(filepath.Join(workingDir, file))
        cfgs = append(cfgs, cfg)
    }

    cfg := cfgs[0]
    for i, c := range cfgs {
        if i == 0 { continue }
        cfg = Join(&cfg, &c)
    }

    var mode FileMode = Many
    if cfg.FileModeString == "one" {
        mode = One
    }
    cfg.FileMode = mode

    return cfg
}

func Join(dest, src *Configuration) Configuration {
    i := reflecthelper.Join(dest, src)
    cfgptr := i.(*Configuration)

    return *cfgptr
}