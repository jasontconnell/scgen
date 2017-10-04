package main

import (
    "fmt"
    "scgen/conf"
    "os"
    "scgen/processor"
    "time"
    "flag"
)

var configFile string

func init(){
    flag.StringVar(&configFile, "c", "", "Configuration filename(s). Use this to run multiple scgen tasks on the same database but targeting different paths and outputs. Can be CSV (order matters)")
}

func main(){
    flag.Parse()
    start := time.Now()
    wd,_ := os.Getwd()

    if configFile == "" {
        flag.PrintDefaults()
        os.Exit(0)
        return
    }

    cfg := conf.LoadConfigs(wd, configFile)

    processor := processor.Processor{ Config: cfg }
    result := processor.Process()

    fmt.Println("Finished generating code in", time.Since(start))
    fmt.Printf("Templates read: %v   Templates processed: %v    Items Read: %v   Items Serialized: %v", result.TemplatesRead, result.TemplatesProcessed, result.ItemsRead, result.ItemsSerialized)
}