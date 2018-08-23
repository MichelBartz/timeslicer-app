package pkg

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"path"
	"sync"

	fixedwidth "github.com/ianlopshire/go-fixedwidth"
)

// DbSyncMessage represents a store row to be saved to disk
type DbSyncMessage struct {
	pk  string
	row *bytes.Buffer
}

// Index is represent an individual index entry
type Index struct {
	Pk         string `fixed:"1,255"`   // primary key value
	RowPos     int64  `fixed:"256,276"` // row position in .db file
	RowByteLen int    `fixed:"277,297"` // row byte length
	IPos       int64  `fixed:"298,318"` // index position in .index file
}

// DbIndex represent a database primary index
type DbIndex = map[string]Index

// FileStore represents the disk storage for our data
type FileStore struct {
	dbDir        string
	fileHandler  *os.File
	indexHandler *os.File
	Connected    bool
	err          error
	message      chan DbSyncMessage
	index        DbIndex
	maxIndex     int64
	mux          sync.Mutex
}

// NewFileStore creates and returns a new FileStore type
func NewFileStore(storeName string) *FileStore {
	fileStore := &FileStore{}

	user, err := user.Current()
	fileStore.err = err

	dbDir, err := initStoreDir(user.HomeDir)
	fileStore.err = err
	fileStore.dbDir = dbDir

	fileStore.index = make(DbIndex)

	fileStore.Connect(storeName)

	fileStore.message = make(chan DbSyncMessage)

	return fileStore
}

// Close terminates the database and its handlers
func (fs *FileStore) Close() {
	fs.fileHandler.Close()
	fs.indexHandler.Close()
	close(fs.message)
}

// Connect loads the database, creates it if non existant
func (fs *FileStore) Connect(dbName string) {
	if fs.err != nil {
		return
	}
	log.Printf("Connecting to '%s'", dbName)

	pathToDb := path.Join(fs.dbDir, fmt.Sprintf("%s.db", dbName))
	file, err := os.OpenFile(pathToDb, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		fs.err = err
		return
	}
	fs.fileHandler = file
	fs.Connected = true
	log.Print("Connected")

	pathToIndexFile := path.Join(fs.dbDir, fmt.Sprintf("%s.index", dbName))

	// Open a handler to the index file for later usage
	indexFile, err := os.OpenFile(pathToIndexFile, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		fs.err = err
		return
	}
	fs.indexHandler = indexFile

	// New version
	data, err := ioutil.ReadFile(pathToIndexFile)
	if err != nil {
		fs.err = err
		return
	}

	var index []Index
	err = fixedwidth.Unmarshal(data, &index)
	if err != nil {
		fs.err = err
		return
	}

	for _, entry := range index {
		fs.index[entry.Pk] = entry
	}

	rebuildFromIndex(fs)

	go fs.DoSync()
}

// DoSync saves the store to the filestore
func (fs *FileStore) DoSync() {
	if fs.err != nil {
		return
	}

	for {
		toSync := <-fs.message
		// Primary key should be enforced as a 255 char string maximum
		fmt.Printf("Saving row: %s\n", toSync.pk)
		// Regardless of update or insert record we append the "new" row at the end of the file
		// We'll update the index if it was an update
		info, err := fs.fileHandler.Stat()
		if err != nil {
			log.Fatal("An error occured adding to file index", err)
		}
		index := Index{
			Pk:         toSync.pk,
			RowPos:     info.Size(),
			RowByteLen: toSync.row.Len(),
		}
		doInsertAt(fs, index, toSync.row)
		if oldIndex, ok := fs.index[toSync.pk]; ok {
			log.Printf("Performing index update at pos: %d...\n", oldIndex.IPos)
			index.IPos = oldIndex.IPos
		} else {
			log.Print("Peforming index insert...\n")
			info, err := fs.indexHandler.Stat()
			if err != nil {
				fs.err = err
				return
			}

			index.IPos = info.Size()
		}
		updateIndex(fs, index)
	}
}

// Get retrieves a record by primary key from the file store
func (fs *FileStore) Get(pk string) (*DbSyncMessage, error) {
	if index, ok := fs.index[pk]; ok {
		var row = make([]byte, index.RowByteLen)

		fs.mux.Lock()
		fs.fileHandler.ReadAt(row, index.RowPos)
		fs.mux.Unlock()

		buffer := bytes.NewBuffer(row)
		return &DbSyncMessage{
			pk:  pk,
			row: buffer,
		}, nil
	}

	return &DbSyncMessage{}, fmt.Errorf("Cannot find record with primary key: %s", pk)
}

// Internals

func rebuildFromIndex(fs *FileStore) {
	// @ToDo: Do rebuild from index
	// Weird? Icky? Unsure
}

func doInsertAt(fs *FileStore, index Index, row *bytes.Buffer) {
	fs.mux.Lock()
	defer fs.mux.Unlock()

	_, err := fs.fileHandler.WriteAt(row.Bytes(), index.RowPos)
	if err != nil {
		fs.err = err
	}
}

func updateIndex(fs *FileStore, index Index) {
	fs.mux.Lock()
	defer fs.mux.Unlock()
	fs.index[index.Pk] = index

	indexEntry, err := fixedwidth.Marshal(index)
	if err != nil {
		fs.err = err
		return
	}
	log.Printf("Seeking at %d", index.IPos)
	if _, err := fs.indexHandler.Seek(index.IPos, 0); err != nil {
		fs.err = err
		return
	}
	if _, err := fs.indexHandler.WriteAt([]byte(indexEntry), index.IPos); err != nil {
		fs.err = err
		return
	}
	if err := fs.indexHandler.Sync(); err != nil {
		fs.err = err
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
