package pkg

import (
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
}

// SetSlice sets the slice value for a given key
func (t *TimeSlicerStore) SetSlice(key, slice, activity string) bool {
	if ds, ok := t.memoryStore[key]; ok {
		if _, ok := ds[slice]; ok {
			ds[slice] = activity
			t.mux.Lock()
			defer t.mux.Unlock()
			t.memoryStore[key] = ds
			return true
		}
	}
	return false
}
