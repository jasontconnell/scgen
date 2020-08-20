package processor

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/jasontconnell/scgen/conf"
	"github.com/jasontconnell/scgen/model"
)

type TemplateData struct {
	Templates []*model.Template
}

var fns = template.FuncMap{
	"plus1": func(x int) int {
		return x + 1
	},
	"lower": func(x string) string {
		return strings.ToLower(x)
	},
}

func generate(cfg conf.Configuration, templates []*model.Template) error {
	err := os.RemoveAll(cfg.OutputPath)

	if err != nil {
		return fmt.Errorf("couldn't remove files %s: %w", cfg.OutputPath, err)
	}

	var outputPath string
	if filepath.Ext(cfg.OutputPath) == "" {
		outputPath = cfg.OutputPath
	} else {
		outputPath = filepath.Dir(cfg.OutputPath)
	}

	err = os.MkdirAll(outputPath, os.ModePerm)
	if err != nil {
		return fmt.Errorf("couldn't make output path %s: %w", outputPath, err)
	}

	list := []*model.Template{}
	for _, t := range templates {
		if t.Generate {
			list = append(list, t)
		}
	}

	var perr error
	if cfg.FileMode == conf.One {
		perr = processOne(cfg, list)
	} else {
		perr = processMany(cfg, list)
	}

	return perr
}

func processInlineTemplate(code string, tmpl *model.Template) (string, error) {
	var value string

	ftmpl := template.New("Template")
	ftmp, err := ftmpl.Parse(code)
	if err != nil {
		return "", fmt.Errorf("parsing string %s: %w", code, err)
	}

	fbuf := new(bytes.Buffer)
	err = ftmp.Execute(fbuf, tmpl)
	if err != nil {
		return "", fmt.Errorf("executing template inline %s %s: %w", code, tmpl.Name, err)
	}

	value = string(fbuf.Bytes())
	return value, nil
}

func processOne(cfg conf.Configuration, templates []*model.Template) error {
	tmpl, err := template.New(cfg.CodeTemplate).Funcs(fns).ParseFiles(cfg.CodeTemplate)

	if err != nil {
		return fmt.Errorf("error parsing template %s: %w", cfg.CodeTemplate, err)
	}

	buffer := new(bytes.Buffer)
	_, templateName := filepath.Split(cfg.CodeTemplate)
	err = tmpl.ExecuteTemplate(buffer, templateName, TemplateData{Templates: templates})

	if err != nil {
		return fmt.Errorf("error executing template %s: %w", cfg.CodeTemplate, err)
	}

	return writeFile(cfg.OutputPath, buffer.Bytes())
}

func processMany(cfg conf.Configuration, templates []*model.Template) error {
	tmpl, err := template.ParseFiles(cfg.CodeTemplate)

	if err != nil {
		return fmt.Errorf("parsing file %s: %w", cfg.CodeTemplate, err)
	}

	for _, sctemplate := range templates {
		p := sctemplate.Path
		for _, bp := range cfg.BasePaths {
			p = strings.TrimPrefix(p, bp)
		}
		dir, _ := path.Split(p)
		dir = strings.Replace(dir, "/", "\\", -1)
		fullPath := filepath.Join(cfg.OutputPath, dir)

		err = os.MkdirAll(fullPath, os.ModePerm)

		if err != nil {
			return fmt.Errorf("making directory for %s : %w", fullPath, err)
		}

		buffer := new(bytes.Buffer)
		_, templateName := filepath.Split(cfg.CodeTemplate)

		err = tmpl.Funcs(fns).ExecuteTemplate(buffer, templateName, TemplateData{Templates: append([]*model.Template{}, sctemplate)})

		if err != nil {
			return err
		}

		filename := filepath.Join(fullPath, sctemplate.CleanName+"."+cfg.CodeFileExtension)
		if cfg.FilenameTemplate != "" {
			name, err := processInlineTemplate(cfg.FilenameTemplate, sctemplate)
			if err != nil {
				return fmt.Errorf("processing inline template %s: %w", cfg.FilenameTemplate, err)
			}
			filename = filepath.Join(fullPath, name+"."+cfg.CodeFileExtension)
		}

		err = writeFile(filename, buffer.Bytes())
		if err != nil {
			return fmt.Errorf("writing file %s: %w", filename, err)
		}
	}

	return nil
}

func writeFile(path string, bytes []byte) error {
	dir, _ := filepath.Split(path)
	os.MkdirAll(dir, os.ModePerm)
	return ioutil.WriteFile(path, bytes, os.ModePerm)
}
