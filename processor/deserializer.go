package processor

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"scgen/conf"
	"scgen/data"
	"strconv"
	"strings"
)

var itemregex *regexp.Regexp = regexp.MustCompile(`ID: (.*?)\r\nName: (.*?)\r\nTemplateID: (.*?)\r\nParentID: (.*?)\r\nMasterID: (.*?)\n`)
var fieldregex *regexp.Regexp = regexp.MustCompile(`(?s)__FIELD__\r\nID: (.*?)\r\nName: (.*?)\r\nVersion: (.*?)\r\nLanguage: (.*?)\r\nSource: (.*?)\r\n__VALUESTART__\r\n(.*?)\r\n___VALUEEND___\r\n\r\n`)

func getItemsForDeserialization(cfg conf.Configuration) []data.DeserializedItem {
	list := []data.DeserializedItem{}
	filepath.Walk(cfg.SerializationPath, func(path string, info os.FileInfo, err error) error {
		if strings.HasSuffix(path, "."+cfg.SerializationExtension) {
			bytes, _ := ioutil.ReadFile(path)
			contents := string(bytes)
			if itemmatches := itemregex.FindAllStringSubmatch(contents, -1); len(itemmatches) == 1 {
				m := itemmatches[0]
				id := m[1]
				name := m[2]
				template := m[3]
				parent := m[4]
				master := m[5]

				item := data.DeserializedItem{ID: id, TemplateID: template, ParentID: parent, Name: name, MasterID: master, Fields: []data.DeserializedField{}}

				if fieldmatches := fieldregex.FindAllStringSubmatch(contents, -1); len(fieldmatches) > 0 {
					for _, m := range fieldmatches {
						id := m[1]
						name := m[2]
						version, _ := strconv.ParseInt(m[3], 10, 64)
						language := m[4]
						source := m[5]
						value := m[6]

						item.Fields = append(item.Fields, data.DeserializedField{ID: id, Name: name, Version: version, Language: language, Source: source, Value: value})
					}
				}
				list = append(list, item)
			}
		}

		return nil
	})

	return list
}

func filterUpdateItems(filteredMap map[string]*data.Item, serialList []*data.SerializedItem, deserialList []data.DeserializedItem) ([]data.UpdateItem, []data.UpdateField) {
	updateItems := []data.UpdateItem{}
	updateFields := []data.UpdateField{}
	itemMap := make(map[string]*data.SerializedItem)
	fieldMap := make(map[string]data.SerializedField)

	for _, sitem := range serialList {
		itemMap[sitem.Item.ID] = sitem
		for _, field := range sitem.Fields {
			key := fmt.Sprintf("%v_%v", sitem.Item.ID, field.FieldID)
			fieldMap[key] = field
		}
	}

	deserializedItemMap := make(map[string]data.DeserializedItem)
	deserializedFieldMap := make(map[string]data.DeserializedField)

	for _, ditem := range deserialList {
		deserializedItemMap[ditem.ID] = ditem
		for _, dfield := range ditem.Fields {
			key := fmt.Sprintf("%v_%v", ditem.ID, dfield.ID)
			deserializedFieldMap[key] = dfield
		}
	}

	for _, sitem := range serialList {
		_, inFilter := filteredMap[sitem.Item.ID]
		if _, ok := deserializedItemMap[sitem.Item.ID]; !ok && inFilter {
			updateItems = append(updateItems, data.UpdateItem{ID: sitem.Item.ID, Name: sitem.Item.Name, TemplateID: sitem.Item.TemplateID, ParentID: sitem.Item.ParentID, MasterID: sitem.Item.MasterID, UpdateType: data.Delete})

			for _, sfield := range sitem.Fields {
				updateFields = append(updateFields, data.UpdateField{ItemID: sitem.Item.ID, FieldID: sfield.FieldID, Value: sfield.Value, Source: sfield.Source, Version: sfield.Version, Language: sfield.Language, UpdateType: data.Delete})
			}
		}
	}

	for _, ditem := range deserialList {
		if item, ok := itemMap[ditem.ID]; !ok {
			updateItems = append(updateItems, data.UpdateItem{ID: ditem.ID, Name: ditem.Name, TemplateID: ditem.TemplateID, ParentID: ditem.ParentID, MasterID: ditem.MasterID, UpdateType: data.Insert})
			for _, field := range ditem.Fields {
				updateFields = append(updateFields, data.UpdateField{ItemID: ditem.ID, FieldID: field.ID, Value: field.Value, Source: field.Source, Version: field.Version, Language: field.Language, UpdateType: data.Insert})
			}
		} else {
			for _, field := range ditem.Fields {
				key := fmt.Sprintf("%v_%v", ditem.ID, field.ID)
				if existingField, ok := fieldMap[key]; !ok {
					updateFields = append(updateFields, data.UpdateField{ItemID: ditem.ID, FieldID: field.ID, Value: field.Value, Source: field.Source, Version: field.Version, Language: field.Language, UpdateType: data.Insert})
				} else {
					if existingField.Value != field.Value || existingField.Version != field.Version || existingField.Language != field.Language {
						updateFields = append(updateFields, data.UpdateField{ItemID: item.Item.ID, FieldID: field.ID, Value: field.Value, Source: field.Source, Version: field.Version, Language: field.Language, UpdateType: data.Update})
					}
				}
			}

			if item.Item.Name != ditem.Name || item.Item.TemplateID != ditem.TemplateID || item.Item.MasterID != ditem.MasterID || item.Item.ParentID != ditem.ParentID {
				updateItems = append(updateItems, data.UpdateItem{ID: ditem.ID, Name: ditem.Name, TemplateID: ditem.TemplateID, ParentID: ditem.ParentID, MasterID: ditem.MasterID, UpdateType: data.Update})

				//fmt.Println("item found in db, updating", ditem.ID)
			}
		}
	}
	return updateItems, updateFields
}
