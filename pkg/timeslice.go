package pkg

import (
	"log"
	"time"
)

type Slice struct {
	startsAt   time.Time
	activities []string
}

type DaySlicer struct {
	date     time.Time
	interval time.Duration
	slices   []Slice
	err      error
}

func NewDaySlicer() DaySlicer {
	now := time.Now()
	return DaySlicer{
		date: time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()),
	}
}

func (ds *DaySlicer) Create(interval string) {
	var err error
	ds.interval, err = time.ParseDuration(interval)
	if err != nil {
		ds.err = err
	}

	day := time.Duration(24) * time.Hour
	numSlices := int(day / ds.interval)
	log.Printf("Creating %d slices", numSlices)
	for i := 0; i < numSlices; i++ {
		slice := Slice{
			startsAt: ds.date.Add(time.Duration(i) * ds.interval),
		}
		log.Printf("Slice at %02dh%02d", slice.startsAt.Hour(), slice.startsAt.Minute())
		ds.slices = append(ds.slices, slice)
	}
}
