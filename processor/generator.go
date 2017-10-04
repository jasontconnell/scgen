package processor

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"scgen/conf"
	"scgen/data"
	"strings"
	"text/template"
)

type TemplateData struct {
	Templates []*data.Template
}

func generate(cfg conf.Configuration, templates []*data.Template) {
	if cfg.FileMode == conf.One {
		processOne(cfg, templates)
	} else {
		processMany(cfg, templates)
	}
}

func processOne(cfg conf.Configuration, templates []*data.Template) {
	os.Remove(cfg.OutputPath)

	if tmpl, err := template.ParseFiles(cfg.CodeTemplate); err == nil {
		buffer := new(bytes.Buffer)
		terr := tmpl.Execute(buffer, TemplateData{Templates: templates})

		if terr != nil {
			panic(terr)
		}

		writeFile(cfg.OutputPath, buffer.Bytes())
	} else {
		fmt.Println("couldn't parse template", err)
	}
}

func processMany(cfg conf.Configuration, templates []*data.Template) {
	if err := os.RemoveAll(cfg.OutputPath); err == nil {
		if err := os.MkdirAll(cfg.OutputPath, os.ModePerm); err == nil {
			if tmpl, err := template.ParseFiles(cfg.CodeTemplate); err == nil {
				for _, template := range templates {
					p := template.Item.Path
					for _, bp := range cfg.BasePaths {
						p = strings.TrimPrefix(p, bp)
					}
					dir, _ := path.Split(p)
					dir = strings.Replace(dir, "/", "\\", -1)
					fullPath := filepath.Join(cfg.OutputPath, dir)

					if cerr := os.MkdirAll(fullPath, os.ModePerm); cerr == nil {
						buffer := new(bytes.Buffer)
						terr := tmpl.Execute(buffer, TemplateData{Templates: append([]*data.Template{}, template)})

						if terr != nil {
							panic(terr)
						}

						filename := filepath.Join(fullPath, template.Item.CleanName+"."+cfg.CodeFileExtension)
						if err := writeFile(filename, buffer.Bytes()); err != nil {
							fmt.Println("error occurred", err)
						}
					}
				}
			}
		}
	} else {
		fmt.Println("error occurred removing all", err)
	}
}

func writeFile(path string, bytes []byte) error {
	return ioutil.WriteFile(path, bytes, os.ModePerm)
}
