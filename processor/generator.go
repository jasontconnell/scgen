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

var fns = template.FuncMap{
	"plus1": func(x int) int {
		return x + 1
	},
}

func groupTemplates(by string, templates []*data.Template) []*[]*data.Template {
	tmpmap := make(map[string]*[]*data.Template)

	for _, tmpl := range templates {
		var key string
		switch by {
		case "Parent":
			key = tmpl.Item.Parent.Parent.CleanName + "." + tmpl.Item.Parent.CleanName
		default:
			key = tmpl.Item.CleanName
		}

		if list, ok := tmpmap[key]; ok {
			dlist := *list
			dlist = append(dlist, tmpl)
			tmpmap[key] = &dlist
		} else {
			list = &[]*data.Template{tmpl}
			tmpmap[key] = list
		}
	}

	ret := []*[]*data.Template{}

	for _, list := range tmpmap {
		ret = append(ret, list)
	}

	return ret
}

func generate(cfg conf.Configuration, templates []*data.Template) {
	if err := os.RemoveAll(cfg.OutputPath); err != nil {
		panic(err)
	}

	var outputPath string
	if filepath.Ext(cfg.OutputPath) == "" {
		outputPath = cfg.OutputPath
	} else {
		outputPath = filepath.Dir(cfg.OutputPath)
	}

	if err := os.MkdirAll(outputPath, os.ModePerm); err != nil {
		panic(err)
	}

	list := []*data.Template{}
	for _, t := range templates {
		if t.Generate {
			list = append(list, t)
		}
	}
	if cfg.FileMode == conf.One {
		processOne(cfg, list)
	} else {
		if cfg.GroupTemplatesBy == "" {
			processMany(cfg, list)
		} else {
			lists := groupTemplates(cfg.GroupTemplatesBy, list)
			for _, group := range lists {
				processOne(cfg, *group)
			}
		}
	}
}

func processInlineTemplate(code string, tmpl *data.Template) string {
	var value string
	ftmpl := template.New("Template")
	if ftmp, ferr := ftmpl.Parse(code); ferr == nil {
		fbuf := new(bytes.Buffer)
		if fexecerr := ftmp.Execute(fbuf, tmpl); fexecerr == nil {
			value = string(fbuf.Bytes())
		}
	} else {
		fmt.Println("couldn't parse filename template", ferr)
	}

	return value
}

func processOne(cfg conf.Configuration, templates []*data.Template) {
	if tmpl, err := template.New(cfg.CodeTemplate).Funcs(fns).ParseFiles(cfg.CodeTemplate); err == nil {
		buffer := new(bytes.Buffer)
		_, templateName := filepath.Split(cfg.CodeTemplate)
		terr := tmpl.ExecuteTemplate(buffer, templateName, TemplateData{Templates: templates})

		if terr != nil {
			panic(terr)
		}

		outputPath := cfg.OutputPath
		if cfg.FilenameTemplate != "" {
			fname := processInlineTemplate(cfg.FilenameTemplate, templates[0]) + "." + cfg.CodeFileExtension
			outputPath = filepath.Join(outputPath, fname)
		}

		writeFile(outputPath, buffer.Bytes())
	} else {
		fmt.Println(err)
	}
}

func processMany(cfg conf.Configuration, templates []*data.Template) {
	if tmpl, err := template.ParseFiles(cfg.CodeTemplate); err == nil {
		for _, sctemplate := range templates {
			p := sctemplate.Item.Path
			for _, bp := range cfg.BasePaths {
				p = strings.TrimPrefix(p, bp)
			}
			dir, _ := path.Split(p)
			dir = strings.Replace(dir, "/", "\\", -1)
			fullPath := filepath.Join(cfg.OutputPath, dir)

			if cerr := os.MkdirAll(fullPath, os.ModePerm); cerr == nil {
				buffer := new(bytes.Buffer)
				_, templateName := filepath.Split(cfg.CodeTemplate)

				terr := tmpl.Funcs(fns).ExecuteTemplate(buffer, templateName, TemplateData{Templates: append([]*data.Template{}, sctemplate)})

				if terr != nil {
					panic(terr)
				}

				filename := filepath.Join(fullPath, sctemplate.Item.CleanName+"."+cfg.CodeFileExtension)
				if cfg.FilenameTemplate != "" {
					name := processInlineTemplate(cfg.FilenameTemplate, sctemplate)
					filename = filepath.Join(fullPath, name+"."+cfg.CodeFileExtension)
				}

				if err := writeFile(filename, buffer.Bytes()); err != nil {
					fmt.Println("error occurred", err)
				}
			}
		}
	} else {
		fmt.Println(err)
	}
}

func writeFile(path string, bytes []byte) error {
	dir, _ := filepath.Split(path)
	os.MkdirAll(dir, os.ModePerm)
	return ioutil.WriteFile(path, bytes, os.ModePerm)
}
