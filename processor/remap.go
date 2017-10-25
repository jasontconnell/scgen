package processor

import (
	"regexp"
	"scgen/conf"
	"scgen/data"
	"strings"
)

var guidReg *regexp.Regexp = regexp.MustCompile("[a-zA-Z0-9\\-]{32,36}")

func processRemap(cfg conf.Configuration, itemMap map[string]*data.Item) map[string]string {
	originalItems := make(map[string]*data.Item)
	clonedItems := make(map[string]*data.Item)

	for _, setting := range cfg.RemapSettings {
		instOriginalItems := getItems(setting.OriginalPath, itemMap)
		for _, item := range instOriginalItems {
			originalItems[item.Name] = item
		}

		instClonedItems := getItems(setting.ClonedPath, itemMap)
		for _, item := range instClonedItems {
			clonedItems[item.Name] = item
		}
	}

	updateMap := make(map[string]string)
	for k, clonedItem := range clonedItems {
		for _, setting := range cfg.RemapSettings {
			newK := strings.Replace(k, setting.ClonedPrefix, setting.OriginalPrefix, -1)
			newK = strings.Trim(newK, " ")
			newK2 := strings.Replace(k, setting.ClonedPrefix, "", -1)
			newK2 = strings.Trim(newK2, " ")

			if originalItem, ok := originalItems[newK]; ok {
				updateMap[originalItem.ID] = clonedItem.ID
				// delete(originalItems, originalItem.Name)
				// delete(clonedItems, clonedItem.Name)
			} else if originalItem, ok := originalItems[newK2]; ok {
				updateMap[originalItem.ID] = clonedItem.ID
				// delete(originalItems, originalItem.Name)
				// delete(clonedItems, clonedItem.Name)
			}
		}
	}

	return updateMap
}

func formatUUID(original, updated string) string {
	if len(original) == len(updated) {
		return updated
	}

	if len(original) < len(updated) {
		return strings.Replace(updated, "-", "", -1)
	}

	return getUUIDKey(updated)
}

func getUUIDKey(uuid string) string {
	if len(uuid) == 36 {
		return strings.ToUpper(uuid)
	}
	key := string(uuid[:8]) + "-" + string(uuid[8:12]) + "-" + string(uuid[12:16]) + "-" + string(uuid[16:20]) + "-" + string(uuid[20:])
	return strings.ToUpper(key)
}

func replaceValues(cfg conf.Configuration, items []*data.SerializedItem, updateMap map[string]string) ([]data.UpdateItem, []data.UpdateField) {
	updateItems := []data.UpdateItem{}
	updateFields := []data.UpdateField{}
	for _, item := range items {
		updateItem := data.UpdateItem{UpdateType: data.Ignore, ID: item.Item.ID, Name: item.Item.Name, ParentID: item.Item.ParentID, MasterID: item.Item.MasterID}
		for _, field := range item.Fields {
			updateField := data.UpdateField{UpdateType: data.Ignore, Value: field.Value, ItemID: item.Item.ID, FieldID: field.FieldID, Source: field.Source, Version: field.Version, Language: field.Language}
			guids := guidReg.FindAllStringSubmatch(field.Value, -1)
			for _, guid := range guids {
				key := getUUIDKey(guid[0])
				if id, ok := updateMap[key]; ok {
					formatted := formatUUID(guid[0], id)
					updateField.UpdateType = data.Update
					updateField.Value = strings.Replace(updateField.Value, guid[0], formatted, -1)
				}
			}

			if fieldId, ok := updateMap[field.FieldID]; ok {
				updateField.FieldID = fieldId
			}

			updateFields = append(updateFields, updateField)
		}

		if templateID, ok := updateMap[item.Item.TemplateID]; ok {
			updateItem.UpdateType = data.Update
			updateItem.TemplateID = templateID
		}

		updateItems = append(updateItems, updateItem)
	}
	return updateItems, updateFields
}

func getItems(path string, itemMap map[string]*data.Item) []*data.Item {
	list := []*data.Item{}
	for _, item := range itemMap {
		if strings.HasPrefix(item.Path, path) {
			list = append(list, item)
		}
	}
	return list
}

func filterRemap(cfg conf.Configuration, dataList []*data.SerializedItem) []*data.SerializedItem {
	list := []*data.SerializedItem{}
	for _, item := range dataList {
		if strings.HasPrefix(item.Item.Path, cfg.RemapApplyPath) {
			list = append(list, item)
		}
	}

	return list
}
