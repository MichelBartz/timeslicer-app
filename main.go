package main

import ts "github.com/MichelBartz/timeslicer-app/pkg"

func main() {
	webserver := ts.NewTimeSlicerWebServer()
	webserver.Start(8080)
}
