package processor

import (
    "fmt"
    "scgen/conf"
)

type Processor struct {
    Config conf.Configuration
}

func (p Processor) Process(itemPath, namespace string){
    if items, err := getItems(p.Config); err == nil {
        root,itemMap,_ := buildTree(items, p.Config.TemplateID, p.Config.TemplateFolderID, p.Config.TemplateFieldID, p.Config.TemplateSectionID)
        fmt.Println(root,len(itemMap))

        templates := mapTemplatesAndFields(p.Config, root)

        fmt.Println(len(templates))
    } else {
        fmt.Println("error occurred", err)
    }
}