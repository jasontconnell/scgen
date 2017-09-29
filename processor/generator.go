package processor

import (
    "fmt"
    "scgen/conf"
    "scgen/data"
    "text/template"
    "bytes"
    "os"
    "bufio"
)

type TemplateData struct {
    Templates []*data.Template
}

func generate (cfg conf.Configuration, templates []*data.Template) {
    fmt.Println("generating code")

    if cfg.FileMode == conf.One {
        processOne(cfg, templates)
    } else {
        processMany(cfg, templates)
    }
}

func processOne(cfg conf.Configuration, templates []*data.Template) {
    os.Remove(cfg.OutputPath)

    if tmpl, err := template.ParseFiles(cfg.OneTemplate); err == nil {
        buffer := new(bytes.Buffer)
        terr := tmpl.Execute(buffer, TemplateData{ Templates: templates } )

        if terr != nil {
            panic(terr)
        }

        if file, err := os.OpenFile(cfg.OutputPath, os.O_CREATE, 0755); err == nil {
            w := bufio.NewWriter(file)
            w.Write(buffer.Bytes())
        } else {
            fmt.Println("Couldn't open output file", err)
        }
    } else {
        fmt.Println("couldn't parse template", err)
    }
}

func processMany(cfg conf.Configuration, templates []*data.Template) {
    
}