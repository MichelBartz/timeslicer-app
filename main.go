package main

import (
	"fmt"
	"log"

	ts "github.com/MichelBartz/timeslicer-app/pkg"
)

func main() {
	var config *ts.TimeslicerConfig
	config = ts.GetConfig()

	timeslicer := ts.NewDaySlicer(config.TimeslicerInterval, config.TimeslicerStart, config.TimeslicerEnd)
	store, err := ts.NewTimeSlicerStore(timeslicer)
	if err != nil {
		log.Fatal(fmt.Errorf("Failed to initial store: %s", err))
	}

	webserver := ts.NewTimeSlicerWebServer(store)
	webserver.Start(config)
}
