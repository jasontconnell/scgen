package processor

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"scgen/conf"
	"scgen/data"
	//"strings"
)

func getSerializeItems(cfg conf.Configuration, itemMap map[string]*data.Item) []*data.SerializedItem {
	list := []*data.SerializedItem{}
	sermap := make(map[string]*data.SerializedItem)
	ids := make(map[string]bool)

	if fieldValues, err := getItemsForSerialization(cfg); err == nil {
		for _, fv := range fieldValues {
			if item, ok := itemMap[fv.ItemID]; ok {
				serfield := data.SerializedField{FieldValueID: fv.FieldValueID, FieldID: fv.FieldID, Name: fv.FieldName, Value: fv.Value, Language: fv.Language, Version: fv.Version, Source: fv.Source}
				if seritem, ok := sermap[fv.ItemID]; ok {
					seritem.Fields = append(seritem.Fields, serfield)
				} else {
					seritem = &data.SerializedItem{Item: item, Fields: []data.SerializedField{serfield}}
					sermap[fv.ItemID] = seritem
					list = append(list, seritem)
				}
			}

			ids[fv.ItemID] = true
		}

		for _, item := range itemMap {
			if _, ok := ids[item.ID]; !ok {
				seritem := &data.SerializedItem{Item: item, Fields: []data.SerializedField{}}
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
	sepend := "___VALUEEND___"

	for _, item := range list {
		//path := item.Item.Path
		//path = strings.Replace(path, "/", "\\", -1)
		path := string(item.Item.ParentID[:1]) + "\\" + string(item.Item.ID[:1])
		dir := filepath.Join(cfg.SerializationPath, path)

		if err := os.MkdirAll(dir, os.ModePerm); err == nil {
			d := fmt.Sprintf("ID: %v\r\nName: %v\r\nTemplateID: %v\r\nParentID: %v\r\nMasterID: %v\r\n\r\n", item.Item.ID, item.Item.Name, item.Item.TemplateID, item.Item.ParentID, item.Item.MasterID)
			for _, f := range item.Fields {
				d += fmt.Sprintf("__FIELD__\r\nID: %v\r\nName: %v\r\nVersion: %v\r\nLanguage: %v\r\nSource: %v\r\n%v\r\n%v\r\n%v\r\n\r\n", f.FieldID, f.Name, f.Version, f.Language, f.Source, sepstart, f.Value, sepend)
			}

			filename := filepath.Join(dir, item.Item.ID+"."+cfg.SerializationExtension)
			ioutil.WriteFile(filename, []byte(d), os.ModePerm)
		}
	}

	return nil
}