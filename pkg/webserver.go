package pkg

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

// TimeSlicerWebServer represent our timeslicer application webserver
type TimeSlicerWebServer struct {
	timeslicerStore Store
	config          *TimeslicerConfig
}

// NewTimeSlicerWebServer creates a new TimeSlicerWebserver
func NewTimeSlicerWebServer(store Store) TimeSlicerWebServer {
	return TimeSlicerWebServer{
		timeslicerStore: store,
	}
}

// Start starts our http webserver for the timeslicer application
func (t *TimeSlicerWebServer) Start(config *TimeslicerConfig) {
	t.config = config

	r := mux.NewRouter()
	r.HandleFunc("/today", t.TodayHandler)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", config.Port), r))
}

// TodayHandler serves the timeslicer-app homepage
func (t *TimeSlicerWebServer) TodayHandler(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	midnight := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	daySlice := NewDaySlicer(t.timeslicerStore, t.config.TimeslicerInterval, t.config.TimeslicerStart, t.config.TimeslicerEnd)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(daySlice.Get(midnight))
}
