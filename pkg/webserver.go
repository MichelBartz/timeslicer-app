package pkg

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

// TimeSlicerApp represent our timeslicer application webserver
type TimeSlicerWebServer struct {
}

// NewTimeSlicerWebServer creates a new TimeSlicerWebserver
func NewTimeSlicerWebServer() TimeSlicerWebServer {
	return TimeSlicerWebServer{}
}

//Start starts our http webserver for the timeslicer application
func (t *TimeSlicerWebServer) Start(port int) {
	r := mux.NewRouter()
	r.HandleFunc("/", HomeHandler)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), r))
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)

	timeslice := NewDaySlicer()
	timeslice.Create("30m")

	fmt.Fprintf(w, "It works \\o/")
}
