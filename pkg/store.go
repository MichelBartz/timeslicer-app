package pkg

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/user"
	"path"
)

// Store is the interface to implement the timeslicer app store
type Store interface {
	Connect()
	IsConnected() bool
	Get(key string) map[string]string
}

// TimeSlicerStore represents the store engine for the timeslicer-app
type TimeSlicerStore struct {
	timeslicer  Slicer
	Connected   bool
	dbDir       string
	fileHandler *os.File
	err         error
}

// NewTimeSlicerStore creates a new store
func NewTimeSlicerStore(slicer Slicer) (*TimeSlicerStore, error) {
	user, err := user.Current()
	if err != nil {
		log.Fatal(fmt.Errorf("Failed to retrieve current user: %s", err))
	}

	dbDir, err := initStoreDir(user.HomeDir)
	if err != nil {
		return nil, errors.New("Failed to initialized timeslicer store")
	}

	return &TimeSlicerStore{
		Connected:  false,
		dbDir:      dbDir,
		timeslicer: slicer,
	}, nil
}

func initStoreDir(homeDir string) (string, error) {
	storeDir := path.Join(homeDir, ".config", "timeslicer")

	if err := os.MkdirAll(storeDir, 0644); err != nil {
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
func (t *TimeSlicerStore) Get(key string) map[string]string {
	slices := make(map[string]string)

	return slices
}
