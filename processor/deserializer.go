package processor

import (
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/jasontconnell/scgen/conf"
	"github.com/jasontconnell/sitecore/api"
	"github.com/jasontconnell/sitecore/data"
)

var itemregex *regexp.Regexp = regexp.MustCompile(`ID: ([a-f0-9\-]{36})\r\nName: (.*?)\r\nTemplateID: ([a-f0-9\-]{36})\r\nParentID: (.*?)\r\nMasterID: ([a-f0-9\-]{36})\r\n\r\n`)
var fieldregex *regexp.Regexp = regexp.MustCompile(`(?s)__FIELD__\r\nID: ([a-f0-9\-]{36})\r\nName: (.*?)\r\nVersion: (.*?)\r\nLanguage: (.*?)\r\nSource: (.*?)\r\n__VALUESTART__\r\n(.*?)\r\n___VALUEEND___\r\n\r\n`)

func getItemsForDeserialization(cfg conf.Configuration) []data.ItemNode {
	list := []data.ItemNode{}
	filepath.Walk(cfg.SerializationPath, func(path string, info os.FileInfo, err error) error {
		if !strings.HasSuffix(path, "."+cfg.SerializationExtension) {
			return nil
		}

		bytes, _ := os.ReadFile(path)
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

					item.AddFieldValue(data.NewFieldValue(id, item.GetId(), name, value, data.Language(language), version, data.FieldSource(source)))
				}
			}
			list = append(list, item)
		}
		return nil
	})

	return list
}
