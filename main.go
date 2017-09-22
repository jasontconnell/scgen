package main

import (
    "scgen/conf"
    "os"
    "path/filepath"
    "scgen/processor"
    "flag"
)

var itemPath string
var namespace string

func init(){
    flag.StringVar(&itemPath, "p", "", "Please provide the template start path")
    flag.StringVar(&namespace, "n", "", "Please provide the base namespace")
}

func main(){
    wd,_ := os.Getwd()

    cfg := conf.LoadConfig(filepath.Join(wd, "config.json"))

    processor := processor.Processor{ Config: cfg }
    processor.Process(itemPath, namespace)
}