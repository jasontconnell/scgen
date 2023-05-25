package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jasontconnell/scgen/conf"
	"github.com/jasontconnell/scgen/process"
)

var templateMessage = `
	******************* FIX IT *******************

	If you are seeing this message, the Go template processor is now a bit more strict about syntax. 
	Particularly the if len gt 0 check for an array. You might see this error:
	executing "template.txt" at <(len $template.[ArrayName]) gt 0>: can't give argument to non-function len $template.[ArrayName]
	The order is [function] param1 param2, there are no infix operators in the template syntax

	Where "gt" is the function.
	So change 
		if (len $template.[ArrayName]) gt 0
	To
		if gt (len $template.[ArrayName]) 0

	ArrayName might be Fields or AllFields or BaseTemplates among others.
	Good luck and happy coding!
	Email jason.connell@herodigital.com with questions.

	********************************************
`

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
	proc := process.Processor{Config: cfg}
	result := proc.Process()

	for _, e := range result.Errors {
		log.Println(fmt.Errorf("An error occurred: %s. %w", e.Error(), e))
	}

	if cfg.Generate && len(result.Errors) > 0 {
		log.Println(templateMessage)
	}

	fmt.Println("Finished process in", time.Since(start))
	fmt.Printf("Items Read: %v   Templates read: %v   Templates processed: %v    Items Serialized: %v   Items Synced: %v   Fields Synced: %v  (Orphans: %v)", result.ItemsRead, result.TemplatesRead, result.TemplatesProcessed, result.ItemsSerialized, result.ItemsDeserialized, result.FieldsDeserialized, result.OrphansCleared)
}
