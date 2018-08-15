package pkg

import (
	"fmt"
	"os"
	"os/user"
	"path"
)

// FileStore represents the disk storage for our data
type FileStore struct {
	dbDir       string
	fileHandler *os.File
	Connected   bool
	err         error
}

// NewFileStore creates and returns a new FileStore type
func NewFileStore(storeName string) *FileStore {
	fileStore := &FileStore{}

	user, err := user.Current()
	fileStore.err = err

	dbDir, err := initStoreDir(user.HomeDir)
	fileStore.err = err
	fileStore.dbDir = dbDir

	fileStore.Connect(storeName)
	return fileStore
}

// Connect creates and load the database
func (fs *FileStore) Connect(dbName string) {
	if fs.err != nil {
		return
	}

	pathToDb := path.Join(fs.dbDir, fmt.Sprintf("%s.db", dbName))
	file, err := os.OpenFile(pathToDb, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		fs.err = err
	}
	fs.fileHandler = file
	fs.Connected = true
}

// DoSync saves the store to the filestore
func (fs *FileStore) DoSync() {
	if fs.err != nil {
		return
	}
}

func initStoreDir(homeDir string) (string, error) {
	storeDir := path.Join(homeDir, ".config", "timeslicer")

	if err := os.MkdirAll(storeDir, 0755); err != nil {
		return "", fmt.Errorf("Could not access store file at %s, %s", storeDir, err)
	}

	return storeDir, nil
}
