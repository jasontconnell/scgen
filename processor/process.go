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

	Error error
}

func (p Processor) Process() ProcessResults {
	results := ProcessResults{}

	if p.Config.Generate {
		tnodes, err := api.LoadTemplates(p.Config.ConnectionString)

		if err != nil {
			results.Error = err
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

	items, err := api.LoadItems(p.Config.ConnectionString)
	if err != nil {
		results.Error = err
		return results
	}

	root, itemMap := api.LoadItemMap(items)
	if root == nil {
		results.Error = fmt.Errorf("No root could be found.")
		return results
	}

	results.ItemsRead = len(itemMap)
	filteredMap := api.FilterItemMap(itemMap, p.Config.BasePaths)

	serialList, err := getSerializeItems(p.Config, filteredMap)
	if err != nil {
		results.Error = err
		return results
	}

	if p.Config.Serialize {
		fmt.Println("Serializing items")
		serializeItems(p.Config, serialList)
		results.ItemsSerialized = len(serialList)
	}

	if p.Config.Deserialize {
		fmt.Println("Getting items for deserialization")
		deserializeItems := getItemsForDeserialization(p.Config)
		updateItems, updateFields := api.BuildUpdateItems(filteredMap, serialList, deserializeItems, true)
		results.ItemsDeserialized = len(updateItems)
		results.FieldsDeserialized = len(updateFields)
		api.Update(p.Config.ConnectionString, updateItems, updateFields)
		results.OrphansCleared = api.CleanOrphanedItems(p.Config.ConnectionString)
	}

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
