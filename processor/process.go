package processor

import (
    "fmt"
    "scgen/conf"
)

type Processor struct {
    Config conf.Configuration
}

func (p Processor) Process(){
    if items, err := getItems(p.Config); err == nil {
        root,_,_ := buildTree(items, p.Config.TemplateID, p.Config.TemplateFolderID, p.Config.TemplateFieldID, p.Config.TemplateSectionID)

        templates := mapTemplatesAndFields(p.Config, root)
        templates = filterTemplates(p.Config, templates)
        mapBaseTemplates(templates)

        updateTemplateNamespaces(p.Config, templates)
        updateReferencedNamespaces(p.Config, templates)
        generate(p.Config, templates)
    } else {
        fmt.Println("error occurred", err)
    }
}