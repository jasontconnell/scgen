package processor

import (
	"fmt"
	"github.com/jasontconnell/sitecore/api"
	"github.com/jasontconnell/sitecore/data"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"scgen/conf"
	"strconv"
	"strings"
)

var itemregex *regexp.Regexp = regexp.MustCompile(`ID: ([a-f0-9\-]{36})\r\nName: (.*?)\r\nTemplateID: ([a-f0-9\-]{36})\r\nParentID: (.*?)\r\nMasterID: ([a-f0-9\-]{36})\r\n\r\n`)
var fieldregex *regexp.Regexp = regexp.MustCompile(`(?s)__FIELD__\r\nID: ([a-f0-9\-]{36})\r\nName: (.*?)\r\nVersion: (.*?)\r\nLanguage: (.*?)\r\nSource: (.*?)\r\n__VALUESTART__\r\n(.*?)\r\n___VALUEEND___\r\n\r\n`)

func getItemsForDeserialization(cfg conf.Configuration) []data.ItemNode {
	list := []data.ItemNode{}
	filepath.Walk(cfg.SerializationPath, func(path string, info os.FileInfo, err error) error {
		if !strings.HasSuffix(path, "."+cfg.SerializationExtension) {
			return nil
		}

		bytes, _ := ioutil.ReadFile(path)
		contents := string(bytes)
		if itemmatches := itemregex.FindAllStringSubmatch(contents, -1); len(itemmatches) == 1 {
			m := itemmatches[0]
			id := api.MustParseUUID(m[1])
			name := m[2]
			template := api.MustParseUUID(m[3])
			parent := api.MustParseUUID(m[4])
			master := api.MustParseUUID(m[5])

			item := data.NewItemNode(id, name, template, parent, master)

			if fieldmatches := fieldregex.FindAllStringSubmatch(contents, -1); len(fieldmatches) > 0 {
				for _, m := range fieldmatches {
					id := api.MustParseUUID(m[1])
					name := m[2]
					version, _ := strconv.ParseInt(m[3], 10, 64)
					language := m[4]
					source := m[5]
					value := m[6]

					item.AddFieldValue(data.NewFieldValue(id, item.GetId(), name, value, language, version, source))
				}
			}
			list = append(list, item)
		}
		return nil
	})

	return list
}

func filterUpdateItems(filteredMap data.ItemMap, serialList []data.ItemNode, deserialList []data.ItemNode) ([]data.UpdateItem, []data.UpdateField) {
	updateItems := []data.UpdateItem{}
	updateFields := []data.UpdateField{}
	itemMap := make(data.ItemMap)
	fieldMap := make(map[string]data.FieldValueNode)

	for _, sitem := range serialList {
		itemMap[sitem.GetId()] = sitem
		for _, field := range sitem.GetFieldValues() {
			key := getFieldKey(sitem, field)
			fieldMap[key] = field
		}
	}

	deserializedItemMap := make(data.ItemMap)
	deserializedFieldMap := make(map[string]data.FieldValueNode)

	for _, ditem := range deserialList {
		deserializedItemMap[ditem.GetId()] = ditem
		for _, dfield := range ditem.GetFieldValues() {
			key := getFieldKey(ditem, dfield)
			deserializedFieldMap[key] = dfield
		}
	}

	for _, sitem := range serialList {
		_, inFilter := filteredMap[sitem.GetId()]
		if _, ok := deserializedItemMap[sitem.GetId()]; !ok && inFilter {
			updateItems = append(updateItems, data.UpdateItemFromItemNode(sitem, data.Delete))

			for _, sfield := range sitem.GetFieldValues() {
				updateFields = append(updateFields, data.UpdateFieldFromFieldValue(sfield, data.Delete))
			}
		}
	}

	for _, ditem := range deserialList {
		if item, ok := itemMap[ditem.GetId()]; !ok {
			updateItems = append(updateItems, data.UpdateItemFromItemNode(ditem, data.Insert))
			for _, field := range ditem.GetFieldValues() {
				updateFields = append(updateFields, data.UpdateFieldFromFieldValue(field, data.Insert))
			}
		} else {
			for _, field := range ditem.GetFieldValues() {
				key := getFieldKey(ditem, field)
				if existingField, ok := fieldMap[key]; !ok {
					updateFields = append(updateFields, data.UpdateFieldFromFieldValue(existingField, data.Insert))
				} else {
					if existingField.GetValue() != field.GetValue() || existingField.GetVersion() != field.GetVersion() || existingField.GetLanguage() != field.GetLanguage() {
						updateFields = append(updateFields, data.UpdateFieldFromFieldValue(field, data.Update))
					}
				}
			}

			if item.GetName() != ditem.GetName() || item.GetTemplateId() != ditem.GetTemplateId() || item.GetMasterId() != ditem.GetMasterId() || item.GetParentId() != ditem.GetParentId() {
				updateItems = append(updateItems, data.UpdateItemFromItemNode(item, data.Update))
			}
		}
	}
	return updateItems, updateFields
}

func getFieldKey(item data.ItemNode, fv data.FieldValueNode) string {
	return fmt.Sprintf("%v_%v_%v", item.GetId(), fv.GetFieldId(), fv.GetLanguage())
}
