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
    if fieldValues, err := getItemsForSerialization(cfg); err == nil {
        for _, fv := range fieldValues {
            if item, ok := itemMap[fv.ID]; ok {
                serfield := data.SerializedField { ID: fv.FieldID, Name: fv.FieldName, Value: fv.Value }
                if seritem, ok := sermap[fv.ID]; ok {
                    seritem.Fields = append(seritem.Fields, serfield)
                } else {
                    seritem = &data.SerializedItem{ Item: item, Fields: []data.SerializedField{ serfield }  }
                    sermap[fv.ID] = seritem
                    list = append(list, seritem)
                }
            }
        }
    } else {
        fmt.Println(err)
    }

    return list
}

func serializeItems(cfg conf.Configuration, list []*data.SerializedItem) error {
    os.RemoveAll(cfg.SerializationOutputPath)
    sepstart := "__VALUESTART__"
    sepend :=   "___VALUEEND___"

    for _, item := range list {
        path := item.Item.Path
        path = strings.Replace(path, "/", "\\", -1)
        dir := filepath.Join(cfg.SerializationOutputPath, path)

        if err := os.MkdirAll(dir, os.ModePerm); err == nil {
            d := fmt.Sprintf("Name: %v\nTemplateID: %v\nParentID: %v\n\n__FIELDS__\n", item.Item.Name, item.Item.TemplateID, item.Item.ParentID)
            for _, f := range item.Fields {
                d += fmt.Sprintf("ID: %v\nName: %v\n%v\n%v\n%v\n\n", f.ID, f.Name, sepstart, f.Value, sepend)
            }

            filename := filepath.Join(dir, item.Item.ID + ".txt")
            ioutil.WriteFile(filename, []byte(d), os.ModePerm)    
        }
    }

    return nil
}