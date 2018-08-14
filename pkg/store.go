package pkg

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/user"
	"path"
	"sync"
)

// Store is the interface to implement the timeslicer app store
type Store interface {
	Connect(db string)
	Get(key string) Slices
	Set(key string, slices Slices)
	SetSlice(key string, slice string, activity string) bool
}

// TimeSlicerStore represents the store engine for the timeslicer-app
type TimeSlicerStore struct {
	Connected   bool
	dbDir       string
	fileHandler *os.File
	memoryStore map[string]Slices
	mux         sync.Mutex
	err         error
}

// NewTimeSlicerStore creates a new store
func NewTimeSlicerStore() (*TimeSlicerStore, error) {
	user, err := user.Current()
	if err != nil {
		log.Fatal(fmt.Errorf("Failed to retrieve current user: %s", err))
	}

	dbDir, err := initStoreDir(user.HomeDir)
	if err != nil {
		return nil, errors.New("Failed to initialized timeslicer store")
	}

	return &TimeSlicerStore{
		Connected:   false,
		dbDir:       dbDir,
		memoryStore: make(map[string]Slices),
	}, nil
}

func initStoreDir(homeDir string) (string, error) {
	storeDir := path.Join(homeDir, ".config", "timeslicer")

	if err := os.MkdirAll(storeDir, 0755); err != nil {
		return "", fmt.Errorf("Could not access store file at %s, %s", storeDir, err)
	}

	return storeDir, nil
}

// Connect creates the connection with the store disk file
func (t *TimeSlicerStore) Connect(db string) {
	pathToDb := path.Join(t.dbDir, fmt.Sprintf("%s.db", db))
	file, err := os.OpenFile(pathToDb, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		t.err = err
	}
	t.fileHandler = file
	t.Connected = true
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
	t.memoryStore[key] = slices
	t.mux.Unlock()
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
