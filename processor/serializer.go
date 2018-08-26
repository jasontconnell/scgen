package processor

import (
	"fmt"
	"github.com/jasontconnell/scgen/conf"
	"github.com/jasontconnell/sitecore/api"
	"github.com/jasontconnell/sitecore/data"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
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
		for _, f := range item.GetFieldValues() {
			d += fmt.Sprintf("__FIELD__\r\nID: %v\r\nName: %v\r\nVersion: %v\r\nLanguage: %v\r\nSource: %v\r\n%v\r\n%v\r\n%v\r\n\r\n", f.GetFieldId(), f.GetName(), f.GetVersion(), f.GetLanguage(), f.GetSource(), sepstart, f.GetValue(), sepend)
		}

		filename := filepath.Join(dir, item.GetId().String()+"."+cfg.SerializationExtension)
		ioutil.WriteFile(filename, []byte(d), os.ModePerm)
	}

	return nil
}
