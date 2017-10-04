package processor

import (
    "fmt"
    "scgen/conf"
    "scgen/data"
    "strings"
    "os"
    "io/ioutil"
    "path/filepath"
)

func getSerializeItems(cfg conf.Configuration, itemMap map[string]*data.Item) []*data.SerializedItem {
    list := []*data.SerializedItem{}
    sermap := make(map[string]*data.SerializedItem)
    ids := make(map[string]bool)

    if fieldValues, err := getItemsForSerialization(cfg); err == nil {
        for _, fv := range fieldValues {
            if item, ok := itemMap[fv.ItemID]; ok {
                serfield := data.SerializedField { FieldValueID: fv.FieldValueID, FieldID: fv.FieldID, Name: fv.FieldName, Value: fv.Value, Language: fv.Language, Version: fv.Version, Source: fv.Source }
                if seritem, ok := sermap[fv.ItemID]; ok {
                    seritem.Fields = append(seritem.Fields, serfield)
                } else {
                    seritem = &data.SerializedItem{ Item: item, Fields: []data.SerializedField{ serfield }  }
                    sermap[fv.ItemID] = seritem
                    list = append(list, seritem)
                }
            }

            ids[fv.ItemID] = true
        }

        for _, item := range itemMap {
            if _,ok := ids[item.ID]; !ok {
                seritem := &data.SerializedItem{ Item: item, Fields: []data.SerializedField{ }  }
                list = append(list, seritem)
            }
        }
    } else {
        fmt.Println(err)
    }

    return list
}

func serializeItems(cfg conf.Configuration, list []*data.SerializedItem) error {
    os.RemoveAll(cfg.SerializationPath)
    sepstart := "__VALUESTART__"
    sepend :=   "___VALUEEND___"

    for _, item := range list {
        path := item.Item.Path
        path = strings.Replace(path, "/", "\\", -1)
        dir := filepath.Join(cfg.SerializationPath, path)

        if err := os.MkdirAll(dir, os.ModePerm); err == nil {
            d := fmt.Sprintf("ID: %v\nName: %v\nTemplateID: %v\nParentID: %v\nMasterID: %v\n\n", item.Item.ID, item.Item.Name, item.Item.TemplateID, item.Item.ParentID, item.Item.MasterID)
            for _, f := range item.Fields {
                d += fmt.Sprintf("__FIELD__\nID: %v\nName: %v\nVersion: %v\nLanguage: %v\nSource: %v\n%v\n%v\n%v\n\n", f.FieldID, f.Name, f.Version, f.Language, f.Source, sepstart, f.Value, sepend)
            }

            filename := filepath.Join(dir, item.Item.ID + "." + cfg.SerializationExtension)
            ioutil.WriteFile(filename, []byte(d), os.ModePerm)    
        }
    }

    return nil
}