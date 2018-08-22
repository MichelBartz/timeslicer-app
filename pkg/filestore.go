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

// IndexEntryLen is the byte size of an index entry,
// this is a constant value for ease of manipulation
const IndexEntryLen = 1234567890

// Index is represent an individual index entry
type Index struct {
	pk         string `fixed:"1,255"`   // primary key value
	rowPos     int64  `fixed:"256,276"` // row position in .db file
	rowByteLen int    `fixed:"277,297"` // row byte length
	iPos       int64  `fixed:"298,318"` // index position in .index file
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

// icky? Unsure if Golang is happy with "global" declaration at package level, methink not
var storeBuilder *StoreBuilder

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

	storeBuilder = NewStoreBuilder(fileStore)

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
	indexFile, err := os.OpenFile(pathToIndexFile, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0644)
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
		fs.index[entry.pk] = entry
	}

	// @ToDo We'll scan the index for a rebuild at startup

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
		fmt.Printf("Saving row: %s", toSync.pk)
		// Regardless of update or insert record we append the "new" row at the end of the file
		// We'll update the index if it was an update
		info, err := fs.fileHandler.Stat()
		if err != nil {
			log.Fatal("An error occured adding to file index", err)
		}
		index := Index{
			pk:         toSync.pk,
			rowPos:     info.Size(),
			rowByteLen: toSync.row.Len(),
		}
		doInsertAt(fs, index, toSync.row)
		if oldIndex, ok := fs.index[toSync.pk]; ok {
			updateIndex(fs, oldIndex, index)
		} else {
			addToIndex(fs, index)
		}
	}
}

// Internals

func doInsertAt(fs *FileStore, index Index, row bytes.Buffer) {
	fs.mux.Lock()
	defer fs.mux.Unlock()

	_, err := fs.fileHandler.WriteAt(row.Bytes(), index.rowPos)
	if err != nil {
		fs.err = err
	}
}

func addToIndex(fs *FileStore, index Index) {
	fs.mux.Lock()
	defer fs.mux.Unlock()
	fs.index[index.pk] = index

	info, err := fs.indexHandler.Stat()
	if err != nil {
		fs.err = err
		return
	}

	index.iPos = info.Size()

	indexEntry, err := fixedwidth.Marshal(index)
	if err != nil {
		fs.err = err
		return
	}
	if _, err := fs.indexHandler.Write([]byte(indexEntry)); err != nil {
		fs.err = err
		return
	}
	if err := fs.indexHandler.Sync(); err != nil {
		fs.err = err
		return
	}
}

func updateIndex(fs *FileStore, old Index, new Index) {
	fs.mux.Lock()
	defer fs.mux.Unlock()
	fs.index[new.pk] = new

	//@ToDo

}

func initStoreDir(homeDir string) (string, error) {
	storeDir := path.Join(homeDir, ".config", "timeslicer")

	if err := os.MkdirAll(storeDir, 0755); err != nil {
		return "", fmt.Errorf("Could not access store file at %s, %s", storeDir, err)
	}

	return storeDir, nil
}
