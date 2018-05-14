package processor

import (
	"fmt"
	"github.com/jasontconnell/sitecore/api"
	"scgen/conf"
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

	if p.Config.Generate {
		tnodes, err := api.LoadTemplates(p.Config.ConnectionString)

		if err != nil {
			fmt.Println("Couldn't get templates:", err)
			return results
		}
		results.TemplatesRead = len(tnodes)
		templates := filterTemplates(p.Config, tnodes)
		results.TemplatesProcessed = len(templates)

		fmt.Println("Generating code")
		generate(p.Config, templates)
	}

	if !p.Config.Serialize && !p.Config.Deserialize && !p.Config.Remap {
		return results
	}

	// root, itemMap := api.LoadItemMap(items)
	// if root == nil {
	// 	fmt.Println("There was a problem getting the items. Exiting.")
	// 	return results
	// }

	// results.ItemsRead = len(itemMap)
	// filteredMap := api.FilterItemMap(itemMap, p.Config.BasePaths)

	// if p.Config.Serialize {
	// 	serialList := getSerializeItems(p.Config, filteredMap)
	// 	fmt.Println("Serializing items")
	// 	serializeItems(p.Config, serialList)
	// 	results.ItemsSerialized = len(serialList)
	// }

	// if p.Config.Deserialize {
	// 	fmt.Println("Getting items for deserialization")
	// 	allList := getSerializeItems(p.Config, itemMap)

	// 	deserializeItems := getItemsForDeserialization(p.Config)
	// 	fmt.Println(len(deserializeItems), "items found on disc")
	// 	updateItems, updateFields := filterUpdateItems(filteredMap, allList, deserializeItems)
	// 	results.ItemsDeserialized = len(updateItems)
	// 	results.FieldsDeserialized = len(updateFields)
	// 	update(p.Config, updateItems, updateFields)
	// 	results.OrphansCleared = cleanOrphanedItems(p.Config)
	// }

	// if p.Config.Remap {
	// 	fmt.Println("Getting items for remap")
	// 	remapList := getSerializeItems(p.Config, filteredMap)

	// 	fmt.Println("Getting remapped ids")
	// 	updateMap := processRemap(p.Config, filteredMap)
	// 	fmt.Println("Filtering for items to remap")
	// 	remapItems := filterRemap(p.Config, remapList)
	// 	fmt.Println("Getting list of update items")
	// 	updateItems, updateFields := replaceValues(p.Config, remapItems, updateMap)
	// 	update(p.Config, updateItems, updateFields)
	// }

	return results
}
