package processor

import (
    "fmt"
    "scgen/conf"
)

type Processor struct {
    Config conf.Configuration
}

type ProcessResults struct {
    TemplatesRead int
    TemplatesProcessed int
}

func (p Processor) Process() ProcessResults {
    results := ProcessResults{}
    if items, err := getItems(p.Config); err == nil {
        root,_,_ := buildTree(items, p.Config.TemplateID, p.Config.TemplateFolderID, p.Config.TemplateFieldID, p.Config.TemplateSectionID)
        if root == nil {
            fmt.Println("There was a problem getting the items. Exiting.")
            return results
        }
        templates := mapTemplatesAndFields(p.Config, root)
        results.TemplatesRead = len(templates)
        templates = filterTemplates(p.Config, templates)
        results.TemplatesProcessed = len(templates)
        mapBaseTemplates(templates)

        updateTemplateNamespaces(p.Config, templates)
        updateReferencedNamespaces(p.Config, templates)
        generate(p.Config, templates)
    } else {
        fmt.Println("error occurred", err)
    }

    return results
}