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
    ItemsRead int
    ItemsSerialized int
}

func (p Processor) Process() ProcessResults {
    results := ProcessResults{}
    fmt.Println("Getting items from database")
    if items, err := getItemsForGeneration(p.Config); err == nil {
        root,itemMap,_ := buildTree(items, p.Config.TemplateID, p.Config.TemplateFolderID, p.Config.TemplateFieldID, p.Config.TemplateSectionID)
        if root == nil {
            fmt.Println("There was a problem getting the items. Exiting.")
            return results
        }

        results.ItemsRead = len(itemMap)

        itemMap = filterItemMap(p.Config, itemMap)

        if p.Config.Generate {
            templates := mapTemplatesAndFields(p.Config, root)
            results.TemplatesRead = len(templates)
            templates = filterTemplates(p.Config, templates)
            results.TemplatesProcessed = len(templates)
            mapBaseTemplates(templates)
            updateTemplateNamespaces(p.Config, templates)
            updateReferencedNamespaces(p.Config, templates)

            fmt.Println("Generating code")
            generate(p.Config, templates)
        }

        if p.Config.Serialize {
            fmt.Println("Serializing items")
            serialList := getSerializeItems(p.Config, itemMap)
            serializeItems(p.Config, serialList)
            results.ItemsSerialized = len(serialList)
        }

    } else {
        fmt.Println("error occurred", err)
    }

    return results
}