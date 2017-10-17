package processor

import (
	"fmt"
	"scgen/conf"
	"scgen/data"
)

type Processor struct {
	Config conf.Configuration
}

type ProcessResults struct {
	TemplatesRead      int
	TemplatesProcessed int
	ItemsRead          int
	ItemsSerialized    int
	ItemsDeserialized  int
	FieldsDeserialized int
	OrphansCleared     int64
}

func (p Processor) Process() ProcessResults {
	results := ProcessResults{}
	fmt.Println("Getting items from database")
	if items, err := getItemsForGeneration(p.Config); err == nil {
		root, itemMap, _ := buildTree(items, p.Config.TemplateID, p.Config.TemplateFolderID, p.Config.TemplateFieldID, p.Config.TemplateSectionID)
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

		var serialList []*data.SerializedItem
		if p.Config.Serialize || p.Config.Deserialize || p.Config.Remap {
			fmt.Println("Getting items for serialization / deserialization / remapping")
			serialList = getSerializeItems(p.Config, itemMap)
		}

		if p.Config.Serialize {
			fmt.Println("Serializing items")
			serializeItems(p.Config, serialList)
			results.ItemsSerialized = len(serialList)
		}

		if p.Config.Deserialize {
			fmt.Println("Getting items for deserialization")
			deserializeItems := getItemsForDeserialization(p.Config)
			fmt.Println(len(deserializeItems), "items for deserialization")
			updateItems, updateFields := filterUpdateItems(serialList, deserializeItems)
			results.ItemsDeserialized = len(updateItems)
			results.FieldsDeserialized = len(updateFields)
			update(p.Config, updateItems, updateFields)
			results.OrphansCleared = cleanOrphanedItems(p.Config)
		}

		if p.Config.Remap {
			fmt.Println("Getting remapped ids")
			updateMap := processRemap(p.Config, itemMap)
			fmt.Println("Filtering for items to remap")
			remapItems := filterRemap(p.Config, serialList)
			fmt.Println("Getting list of update items")
			updateItems, updateFields := replaceValues(p.Config, remapItems, updateMap)
			update(p.Config, updateItems, updateFields)
		}

	} else {
		fmt.Println("error occurred", err)
	}

	return results
}
