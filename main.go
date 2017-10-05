package main

import (
	"flag"
	"fmt"
	"os"
	"scgen/conf"
	"scgen/processor"
	"time"
)

var configFile string

func init() {
	flag.StringVar(&configFile, "c", "", "Configuration filename(s). Use this to run multiple scgen tasks on the same database but targeting different paths and outputs. Can be CSV (order matters)")
}

func main() {
	flag.Parse()
	start := time.Now()
	wd, _ := os.Getwd()

	if configFile == "" {
		flag.PrintDefaults()
		os.Exit(0)
		return
	}

	cfg := conf.LoadConfigs(wd, configFile)

	processor := processor.Processor{Config: cfg}
	result := processor.Process()

	fmt.Println("Finished process in", time.Since(start))
	fmt.Printf("Items Read: %v   Templates read: %v   Templates processed: %v    Items Serialized: %v   Items Synced: %v   Fields Synced: %v  (Orphans: %v)", result.ItemsRead, result.TemplatesRead, result.TemplatesProcessed, result.ItemsSerialized, result.ItemsDeserialized, result.FieldsDeserialized, result.OrphansCleared)
}
