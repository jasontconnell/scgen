package processor

import (
    "strings"
    "scgen/data"
    "scgen/conf"
    "path"
    "utility"
    "sort"
)

func buildTree(items []*data.Item, templateID, templateFolderID, templateFieldID, templateSectionID string) (root *data.Item, itemMap map[string]*data.Item, err error) {
    itemMap = make(map[string]*data.Item)
    for _, item := range items {
        itemMap[item.ID] = item
    }

    root = nil
    for _, item := range itemMap {
        if p, ok := itemMap[item.ParentID]; ok {
            p.Children = append(p.Children, item)
            item.Parent = p
        } else if item.ParentID == "00000000-0000-0000-0000-000000000000" {
            root = item
        }
    }

    if root != nil {
        root.Path = "/" + root.Name
        assignPaths(root)
        assignBaseTemplates(itemMap)
    }
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

func mapBaseTemplates(templates []*data.Template) map[string]*data.Template {
    tmap := make(map[string]*data.Template)
    for _, template := range templates {
        tmap[template.Item.ID] = template
    }

    for _, template := range tmap {
        for _, baseItem := range template.Item.BaseTemplateItems {
            if baseTemplate, ok := tmap[baseItem.ID]; ok {
                template.BaseTemplates = append(template.BaseTemplates, baseTemplate)
            }
        }
    }

    return tmap
}

func mapTemplatesAndFields(cfg conf.Configuration, item *data.Item) []*data.Template {
    templates := []*data.Template{}

    for _, c := range item.Children {
        if c.TemplateID == cfg.TemplateID {
            fields := getFieldsFromTemplate(cfg, c)
            ns := strings.Replace(c.Path, "/", ".", -1)
            ns = strings.Replace(ns, " ", "", -1)
            template := &data.Template{ Item: c,  Fields: fields, Namespace: ns }
            templates = append(templates, template)
        } else {
            ts := mapTemplatesAndFields(cfg, c)
            templates = append(templates, ts...)
        }
    }

    sort.Slice(templates, func(i, j int) bool {
        return templates[i].Item.Path < templates[j].Item.Path
    })

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
    key := strings.ToLower(item.FieldType)
    var fieldType conf.FieldType
    var codeType string
    var ok bool
    if fieldType,ok = cfg.FieldTypeMap[key]; ok {
        codeType = fieldType.CodeType
    } else {
        codeType = cfg.DefaultFieldType
    }
    return data.Field{ Item: item, Name: item.Name, CleanName: item.CleanName, TypeName: item.FieldType, CodeType: codeType, Suffix: fieldType.Suffix }
}

func updateTemplateNamespaces(cfg conf.Configuration, templates []*data.Template){
    for _, template := range templates {
        removeBasePath := template.Item.Path
        for _, basePath := range cfg.BasePaths {
            removeBasePath = strings.TrimPrefix(removeBasePath, basePath)
        }
        removeBasePath, _ = path.Split(removeBasePath)
        nospaces := strings.Replace(strings.TrimSuffix(removeBasePath, "/"), " ", "", -1)
        template.Namespace = strings.Replace(nospaces, "/", ".", -1)
        template.Namespace = cfg.BaseNamespace + strings.Replace(template.Namespace, "/", ".", -1)
    }
}

func updateReferencedNamespaces(cfg conf.Configuration, templates []*data.Template){
    for _, template := range templates {
        template.ReferencedNamespaces = []string{}
        for _, base := range template.BaseTemplates {
            if base.Namespace != template.Namespace && !utility.HasString(template.ReferencedNamespaces, base.Namespace) {
                template.ReferencedNamespaces = append(template.ReferencedNamespaces, base.Namespace)
            }
        }
    }
}

func filterItemMap(cfg conf.Configuration, items map[string]*data.Item) (map[string]*data.Item) {
    filteredMap := make(map[string]*data.Item)
    for _, item := range items {
        include := false
        for _, basePath := range cfg.BasePaths {
            if !include && strings.HasPrefix(item.Path, basePath){
                include = true
                break
            }
        }

        if include {
            filteredMap[item.ID] = item
        }
    }
    return filteredMap
}

func filterTemplates(cfg conf.Configuration, templates []*data.Template) (list []*data.Template) {
    for _, template := range templates {
        include := false
        for _, basePath := range cfg.BasePaths {
            if !include && strings.HasPrefix(template.Item.Path, basePath){
                include = true
                break
            }
        }

        if include {
            list = append(list, template)
        }
    }
    return
}