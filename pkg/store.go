package pkg

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/user"
	"path"
	"time"
)

type Store interface {
	Connect()
	IsConnected() bool
	Get(t time.Time) map[string]string
}

type TimeSlicerStore struct {
	timeslicer  Slicer
	connected   bool
	dbDir       string
	fileHandler *os.File
	err         error
}

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
		connected:  false,
		dbDir:      dbDir,
		timeslicer: slicer,
	}, nil
}

func initStoreDir(homeDir string) (string, error) {
	storeDir := path.Join(homeDir, ".config", "timeslicer")

	if err := os.MkdirAll(storeDir, 0644); err != nil {
		return "", errors.New(fmt.Sprintf("Could not access store file at %s, %s", storeDir, err))
	}

	return storeDir, nil
}

func (t *TimeSlicerStore) Connect(db string) {
	pathToDb := path.Join(t.dbDir, fmt.Sprintf("%s.db", db))
	file, err := os.OpenFile(pathToDb, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		t.err = err
	}
	t.fileHandler = file
	t.connected = true
}

func (t *TimeSlicerStore) IsConnected() bool {
	return t.connected
}

func (t *TimeSlicerStore) Get(day time.Time) map[string]string {
	slices := make(map[string]string)

	return slices
}
