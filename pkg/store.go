package pkg

import (
	"bytes"
	"encoding/gob"
	"sync"
)

// Store is the interface to implement the timeslicer app store
type Store interface {
	Get(key string) Slices
	Set(key string, slices Slices)
	SetSlice(key string, slice string, activity string) bool
}

// TimeSlicerStore represents the store engine for the timeslicer-app
type TimeSlicerStore struct {
	Connected   bool
	fileStore   *FileStore
	memoryStore map[string]Slices
	mux         sync.Mutex
	err         error
}

// NewTimeSlicerStore creates a new store
func NewTimeSlicerStore(storeName string) (*TimeSlicerStore, error) {
	fileStore := NewFileStore(storeName)
	if fileStore.err != nil {
		return nil, fileStore.err
	}

	return &TimeSlicerStore{
		Connected:   false,
		fileStore:   fileStore,
		memoryStore: make(map[string]Slices),
	}, nil
}

// Get returns a key from the store
func (t *TimeSlicerStore) Get(key string) Slices {
	if val, ok := t.memoryStore[key]; ok {
		return val
	}
	return nil
}

// Set creates a new entry in the store
func (t *TimeSlicerStore) Set(key string, slices Slices) {
	t.mux.Lock()
	defer t.mux.Unlock()
	t.memoryStore[key] = slices

	syncToFs(t.fileStore, key, slices)
}

// SetSlice sets the slice value for a given key
func (t *TimeSlicerStore) SetSlice(key, slice, activity string) bool {
	if ds, ok := t.memoryStore[key]; ok {
		if _, ok := ds[slice]; ok {
			ds[slice] = activity
			t.mux.Lock()
			defer t.mux.Unlock()
			t.memoryStore[key] = ds

			if syncToFs(t.fileStore, key, ds) {
				return true
			}
		}
	}
	return false
}

func syncToFs(fs *FileStore, key string, row Slices) bool {
	var rowBuf bytes.Buffer
	enc := gob.NewEncoder(&rowBuf)
	err := enc.Encode(row)
	if err != nil {
		return false
	}
	fs.message <- DbSyncMessage{
		pk:  key,
		row: &rowBuf,
	}
	return true
}
