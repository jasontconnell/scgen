package processor

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/jasontconnell/scgen/conf"
	"github.com/jasontconnell/sitecore/api"
	"github.com/jasontconnell/sitecore/data"
)

func getSerializeItems(cfg conf.Configuration, itemMap data.ItemMap) ([]data.ItemNode, error) {
	list := []data.ItemNode{}
	fieldValues, err := api.LoadFieldsParallel(cfg.ConnectionString, 24)

	if err != nil {
		return nil, err
	}

	api.AssignFieldValues(itemMap, fieldValues)
	for _, v := range itemMap {
		list = append(list, v)
	}

	return list, nil
}

func serializeItems(cfg conf.Configuration, list []data.ItemNode) error {
	os.RemoveAll(cfg.SerializationPath)

	if len(list) > 100 {
		var wg sync.WaitGroup
		wg.Add(6)
		groupSize := len(list)/6 + 1

		for i := 0; i < 6; i++ {
			grp := list[(i * groupSize) : (i+1)*groupSize]
			go func(grplist []data.ItemNode) {
				serializeItemGroup(cfg, grplist)
				wg.Done()
			}(grp)
		}

		wg.Wait()
	} else {
		serializeItemGroup(cfg, list)
	}

	return nil
}

func serializeItemGroup(cfg conf.Configuration, list []data.ItemNode) error {
	sepstart := "__VALUESTART__"
	sepend := "___VALUEEND___"

	ignoreDefault := false
	ignoreFields := make(map[string]bool)
	for _, f := range cfg.SerializationIgnoredFields {
		fn := f
		ignore := true
		if f[0] == '-' {
			ignore = false
			fn = string(f[1:])
			ignoreDefault = true
		}
		ignoreFields[strings.ToLower(fn)] = ignore
	}

	for _, item := range list {
		if item == nil {
			continue
		}

		strid := item.GetId().String()
		path := filepath.Join(string(strid[0]), string(strid[1]))
		dir := filepath.Join(cfg.SerializationPath, path)

		err := os.MkdirAll(dir, os.ModePerm)

		if err != nil {
			return err
		}

		d := fmt.Sprintf("ID: %v\r\nName: %v\r\nTemplateID: %v\r\nParentID: %v\r\nMasterID: %v\r\n\r\n", item.GetId(), item.GetName(), item.GetTemplateId(), item.GetParentId(), item.GetMasterId())

		sorted := item.GetFieldValues()
		sort.Slice(sorted, func(i, j int) bool {
			if sorted[i].GetName() < sorted[j].GetName() {
				return true
			} else if sorted[i].GetLanguage() < sorted[j].GetLanguage() {
				return true
			} else {
				return sorted[i].GetName() == sorted[j].GetName() && sorted[i].GetLanguage() == sorted[j].GetLanguage() && sorted[i].GetVersion() < sorted[j].GetVersion()
			}
		})

		for _, f := range sorted {
			if ignore, ok := ignoreFields[strings.ToLower(f.GetName())]; ok && ignore || (!ok && ignoreDefault) {
				continue
			}

			d += fmt.Sprintf("__FIELD__\r\nID: %v\r\nName: %v\r\nVersion: %v\r\nLanguage: %v\r\nSource: %v\r\n%v\r\n%v\r\n%v\r\n\r\n", f.GetFieldId(), f.GetName(), f.GetVersion(), f.GetLanguage(), f.GetSource(), sepstart, f.GetValue(), sepend)
		}

		filename := filepath.Join(dir, item.GetId().String()+"."+cfg.SerializationExtension)
		ioutil.WriteFile(filename, []byte(d), os.ModePerm)
	}

	return nil
}
