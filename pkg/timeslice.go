package pkg

import (
	"fmt"
	"log"
	"time"
)

// Slice represents a chunk of time with associated activities
type Slice struct {
	startsAt   time.Time
	activities []string
}

// Slicer interface
type Slicer interface {
	Create()
	GetSlices() map[string]string
}

// DaySlicer represents a day sliced using an interval in minutes
type DaySlicer struct {
	date     time.Time
	start    time.Time
	end      time.Time
	interval time.Duration
	slices   []Slice
	err      error
}

// NewDaySlicer creates a new DaySlicer struct to interact with
func NewDaySlicer(interval, start, end string) *DaySlicer {
	now := time.Now()
	midnight := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	startDuration, err := time.ParseDuration(start)
	endDuration, err := time.ParseDuration(end)
	if err != nil {
		log.Fatal(fmt.Errorf("Cannot parse duration for start or end of timeslice: %s", err))
	}
	startTime := midnight.Add(startDuration)
	endTime := midnight.Add(endDuration)

	sliceInterval, err := time.ParseDuration(interval)
	if err != nil {
		log.Fatal(fmt.Errorf("Cannot parse duration for interval: %s", err))
	}

	return &DaySlicer{
		start:    startTime,
		end:      endTime,
		interval: sliceInterval,
		date:     now,
	}
}

// Create slices the day in defined interval
func (ds *DaySlicer) Create() {
	currentSlice := Slice{
		startsAt: ds.start,
	}
	for currentSlice.startsAt.Before(ds.end) || currentSlice.startsAt.Equal(ds.end) {
		ds.slices = append(ds.slices, currentSlice)
		log.Printf("Slice at %s", currentSlice.String())
		currentSlice.startsAt = currentSlice.startsAt.Add(ds.interval)
	}
}

// GetSlices returns all the slices of the given DaySlicer
func (ds *DaySlicer) GetSlices() map[string]string {
	slices := make(map[string]string)

	for _, slice := range ds.slices {
		slices[slice.String()] = slice.GetActivity()
	}

	return slices
}

func (s *Slice) String() string {
	return fmt.Sprintf("%02dh%02d", s.startsAt.Hour(), s.startsAt.Minute())
}

// GetActivity returns the activity of the given Slice
func (s *Slice) GetActivity() string {
	if len(s.activities) == 0 {
		return ""
	}
	return s.activities[len(s.activities)-1]
}
