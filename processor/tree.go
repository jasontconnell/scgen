package processor

import (
    "fmt"
    "strings"
    "scgen/data"
    "scgen/conf"
)

func buildTree(items []*data.Item, templateID, templateFolderID, templateFieldID, templateSectionID string) (root *data.Item, itemMap map[string]*data.Item, err error) {
    itemMap = make(map[string]*data.Item)
    for _, item := range items {
        itemMap[item.ID] = item
    }

    fmt.Println("building tree")

    root = nil
    for _, item := range itemMap {
        if p, ok := itemMap[item.ParentID]; ok {
            p.Children = append(p.Children, item)
            item.Parent = p
        } else if item.ParentID == "11111111-1111-1111-1111-111111111111" {
            root = item
        }
    }

    root.Path = ""
    assignPaths(root)
    assignBaseTemplates(itemMap)
    
    return root, itemMap, nil
}

func assignBaseTemplates(itemMap map[string]*data.Item){
    for _, item := range itemMap {

        ids := strings.Split(item.BaseTemplates, "|")
        if len(ids) > 0 {
            item.BaseTemplateItems = []*data.Item{}
            for _, id := range ids {
                if baseTemplate, ok := itemMap[id]; ok {
                    item.BaseTemplateItems = append(item.BaseTemplateItems, baseTemplate)
                }
            }
        }
    }
}

func assignPaths(root *data.Item){
    for i := 0; i < len(root.Children); i++ {
        root.Children[i].Path = root.Path + "/" + root.Children[i].Name
        assignPaths(root.Children[i])
    }
}

func mapTemplatesAndFields(cfg conf.Configuration, item *data.Item) []data.Template {
    templates := []data.Template{}

    for _, c := range item.Children {
        if c.TemplateID == cfg.TemplateID {
            fields := getFieldsFromTemplate(cfg, c)
            ns := strings.Replace(c.Path, "/", ".", -1)
            ns = strings.Replace(ns, " ", "", -1)
            template := data.Template{ Item: c,  Fields: fields, Namespace: ns }
            templates = append(templates, template)
        } else if c.TemplateID == cfg.TemplateFolderID {
            ts := mapTemplatesAndFields(cfg, c)
            templates = append(templates, ts...)
        }
    }

    return templates
}

func getFieldsFromTemplate(cfg conf.Configuration, item *data.Item) []data.Field {
    fields := []data.Field{}
    for _, c := range item.Children {
        if c.TemplateID == cfg.TemplateSectionID {
            tsfields := getFieldsFromTemplateSection(cfg, c)
            fields = append(fields, tsfields...)
        }
    }
    return fields
}

func getFieldsFromTemplateSection(cfg conf.Configuration, item *data.Item) []data.Field {
    fields := []data.Field{}
    for _, c := range item.Children {
        if c.TemplateID == cfg.TemplateFieldID {
            field := getField(cfg, c)
            fields = append(fields, field)
        }
    }
    return fields
}

func getField(cfg conf.Configuration, item *data.Item) data.Field {
    return data.Field{ Name: item.Name, Type: item.FieldType }
}