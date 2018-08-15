package main

import (
	"fmt"
	"log"

	ts "github.com/MichelBartz/timeslicer-app/pkg"
)

func main() {
	var config *ts.TimeslicerConfig
	config = ts.GetConfig()

	store, err := ts.NewTimeSlicerStore(config.StoreName)
	if err != nil {
		log.Fatal(fmt.Errorf("Failed to initial store: %s", err))
	}

	webserver := ts.NewTimeSlicerWebServer(store)
	webserver.Start(config)
}
