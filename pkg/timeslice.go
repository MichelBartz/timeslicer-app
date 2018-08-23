package pkg

import (
	"fmt"
	"log"
	"time"
)

// Slices is an alias for map[string]string for code readability
type Slices = map[string]string

// Slice represents a chunk of time with associated activities
type Slice struct {
	startsAt time.Time
	activity string
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
	store    Store
	err      error
}

// NewDaySlicer creates a new DaySlicer struct to interact with
func NewDaySlicer(store Store, interval, start, end string) *DaySlicer {
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
		store:    store,
	}
}

// Create slices the day in defined interval
func (ds *DaySlicer) Create() {
	currentSlice := Slice{
		startsAt: ds.start,
	}
	for currentSlice.startsAt.Before(ds.end) || currentSlice.startsAt.Equal(ds.end) {
		ds.slices = append(ds.slices, currentSlice)
		currentSlice.startsAt = currentSlice.startsAt.Add(ds.interval)
	}
}

// GetSlices returns all the slices of the given DaySlicer
func (ds *DaySlicer) GetSlices() Slices {
	slices := make(Slices)

	for _, slice := range ds.slices {
		slices[slice.String()] = slice.activity
	}

	return slices
}

// Get Better but still icky methink.
func (ds *DaySlicer) Get(day time.Time) Slices {
	key := TimeToKey(day)
	slices := ds.store.Get(key)
	if slices == nil {
		log.Printf("Creating day slice for %s", day.String())
		ds.Create()
		slices = ds.GetSlices()
		ds.store.Set(key, slices)
	}

	return slices
}

func (s *Slice) String() string {
	return fmt.Sprintf("%02dh%02d", s.startsAt.Hour(), s.startsAt.Minute())
}

// TimeToKey transforms a given date to a useable timeslice key
func TimeToKey(t time.Time) string {
	date := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())

	return date.Format(time.ANSIC)
}
