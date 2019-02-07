package conf

import (
	"encoding/json"
	"path/filepath"
	"strings"

	"github.com/jasontconnell/conf"
	"github.com/jasontconnell/refhelp"
)

type FileMode int

const (
	One FileMode = iota
	Many
)

type Configuration struct {
	TemplateID        string `json:"template"`
	TemplateFolderID  string `json:"templateFolder"`
	TemplateSectionID string `json:"templateSection"`
	TemplateFieldID   string `json:"templateField"`

	FieldTypes       []FieldType `json:"fieldTypes"`
	FieldTypeMap     map[string]FieldType
	DefaultFieldType string `json:"defaultFieldType"`

	CodeTemplate      string `json:"codeTemplate"`
	CodeFileExtension string `json:"codeFileExtension"`

	SerializationIgnoredFields []string `json:"serializationIgnoredFields"`
	SerializationPath          string   `json:"serializationPath"`
	SerializationExtension     string   `json:"serializationExtension"`

	ConnectionString string `json:"connectionString"`

	Serialize   bool `json:"serialize"`
	Generate    bool `json:"generate"`
	Deserialize bool `json:"deserialize"`

	BasePath         string         `json:"basePath"`
	BasePaths        []string       `json:"basePaths"`
	FileModeString   string         `json:"filemode"`
	OutputPath       string         `json:"outputPath"`
	FilenameTemplate string         `json:"filenameTemplate"`
	GroupTemplatesBy string         `json:"groupTemplatesBy"`
	TemplatePaths    []TemplatePath `json:"templatePaths"`

	Remap          bool            `json:"remap"`
	RemapSettings  []RemapSettings `json:"remapSettings"`
	RemapApplyPath string          `json:"remapApplyPath"`

	// not in config file
	FileMode FileMode
}

type FieldType struct {
	TypeName      string          `json:"typeName"`
	CodeType      string          `json:"codeType"`
	Suffix        string          `json:"suffix"`
	PropertiesRaw json.RawMessage `json:"properties"`
	Properties    FieldTypePropertyMap
}

type FieldTypePropertyMap map[string]string

type TemplatePath struct {
	Path               string `json:"path"`
	Namespace          string `json:"namespace"`
	AlternateNamespace string `json:"alternateNamespace"`
	Ignore             bool   `json:"ignore"`
	StaticNamespace    bool   `json:"staticNamespace"` // give all templates under template path the same namespace
}

type RemapSettings struct {
	OriginalPath   string `json:"originalPath"`
	ClonedPath     string `json:"clonedPath"`
	OriginalPrefix string `json:"originalPrefix"`
	ClonedPrefix   string `json:"clonedPrefix"`

	// future use
	OriginalSuffix string `json:"originalSuffix"`
	ClonedSuffix   string `json:"clonedSuffix"`
}

func LoadConfig(file string) Configuration {
	config := Configuration{}
	conf.LoadConfig(file, &config)

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
		if i == 0 {
			continue
		}
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
	i := refhelp.Join(dest, src)
	cfgptr := i.(*Configuration)

	if len(cfgptr.FieldTypes) > 0 {
		cfgptr.FieldTypeMap = make(map[string]FieldType)
		for _, ft := range cfgptr.FieldTypes {
			key := strings.ToLower(ft.TypeName)
			if ft.PropertiesRaw != nil {
				ex, _ := conf.DecodeRawMessage(ft.PropertiesRaw)
				extra := ex.(map[string]interface{})
				ft.Properties = make(map[string]string)

				for k, v := range extra {
					ft.Properties[k] = v.(string)
				}
			}

			cfgptr.FieldTypeMap[key] = ft
		}
	}

	if len(cfgptr.BasePaths) == 0 && len(cfgptr.BasePath) > 0 {
		cfgptr.BasePaths = append(cfgptr.BasePaths, cfgptr.BasePath)
	}

	return *cfgptr
}
