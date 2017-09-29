package main

import (
    "fmt"
    "scgen/conf"
    "os"
    "path/filepath"
    "scgen/processor"
    "time"
)

func main(){
    start := time.Now()
    wd,_ := os.Getwd()

    cfg := conf.LoadConfig(filepath.Join(wd, "config.json"))
    var mode conf.FileMode = conf.Many
    if cfg.FileModeString == "one" {
        mode = conf.One
    }
    cfg.FileMode = mode

    processor := processor.Processor{ Config: cfg }
    result := processor.Process()

    fmt.Println("Finished generating code in", time.Since(start))
    s := fmt.Sprintf("Templates read: %v  Templates processed: %v", result.TemplatesRead, result.TemplatesProcessed)
    fmt.Println(s)
}