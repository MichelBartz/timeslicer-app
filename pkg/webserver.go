package pkg

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
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
	r.HandleFunc("/dayslice/{timestamp:[0-9]+}", t.DaySliceHandler).Methods("GET")
	r.HandleFunc("/slice", t.SliceHandler).Methods("POST")
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", config.Port), r))
}

// DaySliceHandler is the GET HTTP endpoint handler for /dayslice/timestamp:[0-9]+ to retrieve slices for a given day
func (t *TimeSlicerWebServer) DaySliceHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	response := make(map[string]interface{})

	timestamp, err := strconv.ParseInt(vars["timestamp"], 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		response["error"] = "Invalid timestamp provided."
	}

	daySlice := NewDaySlicer(t.timeslicerStore, t.config.TimeslicerInterval, t.config.TimeslicerStart, t.config.TimeslicerEnd)
	dayTime := time.Unix(timestamp, 0)
	response["date"] = dayTime.String()
	response["slices"] = daySlice.Get(dayTime)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// SliceHandler is the POST HTTP endpoint handler for /slice to add a new activity on a slice
func (t *TimeSlicerWebServer) SliceHandler(w http.ResponseWriter, r *http.Request) {
	// ToDo: Implement details of save slice
	// timestamp := r.FormValue("timestamp")
}
