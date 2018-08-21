package pkg

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"os"
	"os/user"
	"path"
	"strconv"
	"strings"
	"sync"
)

// Index is represent an individual index entry
type Index struct {
	pos     int64
	byteLen int
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

	// Load up index file and warm up in memory index
	pathToIndexFile := path.Join(fs.dbDir, fmt.Sprintf("%s.index", dbName))
	indexFile, err := os.OpenFile(pathToIndexFile, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		fs.err = err
		return
	}
	fs.indexHandler = indexFile
	// Index lines are formatted as such: pk,index;pk,index
	scanner := bufio.NewScanner(fs.indexHandler)
	onSemicolon := func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		for i := 0; i < len(data); i++ {
			if data[i] == ';' {
				return i + 1, data[:i], nil
			}
		}
		return 0, data, bufio.ErrFinalToken
	}
	scanner.Split(onSemicolon)
	for scanner.Scan() {
		index := scanner.Text()
		// Could probably run this concurrently so that processing massive indexes is not an issue ?
		if len(index) > 1 {
			s := strings.Split(index, ",")
			pos, err := strconv.ParseInt(s[1], 10, 64)
			byteLen, err := strconv.ParseInt(s[1], 10, 64)
			if err != nil {
				fs.err = err
				return
			}

			fs.index[s[0]] = Index{
				pos:     pos,
				byteLen: int(byteLen),
			}
		}
	}

	go fs.DoSync()
}

// DoSync saves the store to the filestore
func (fs *FileStore) DoSync() {
	if fs.err != nil {
		return
	}

	for {
		toSync := <-fs.message
		fmt.Printf("Saving row: %s", toSync.pk)
		// Regardless of update or insert record we append the "new" row at the end of the file
		// We'll update the index if it was an update
		info, err := fs.fileHandler.Stat()
		if err != nil {
			log.Fatal("An error occured adding to file index", err)
		}
		index := Index{
			pos:     info.Size(),
			byteLen: toSync.row.Len(),
		}
		doInsertAt(fs, index, toSync.row)
		if oldIndex, ok := fs.index[toSync.pk]; ok {
			updateIndex(fs, oldIndex, index)
		} else {
			addToIndex(fs, toSync.pk, index)
		}
	}
}

// Internals

func doInsertAt(fs *FileStore, index Index, row bytes.Buffer) {
	fs.mux.Lock()
	defer fs.mux.Unlock()

	_, err := fs.fileHandler.WriteAt(row.Bytes(), index.pos)
	if err != nil {
		fs.err = err
	}
}

func addToIndex(fs *FileStore, pk string, index Index) {
	fs.mux.Lock()
	defer fs.mux.Unlock()
	fs.index[pk] = index

	indexEntry := fmt.Sprintf("%s,%d,%d;", pk, index.pos, index.byteLen)
	if _, err := fs.indexHandler.Write([]byte(indexEntry)); err != nil {
		log.Print(err)
		fs.err = err
	}
	if err := fs.indexHandler.Sync(); err != nil {
		fs.err = err
	}
}

func updateIndex(fs *FileStore, old Index, new Index) {

}

func initStoreDir(homeDir string) (string, error) {
	storeDir := path.Join(homeDir, ".config", "timeslicer")

	if err := os.MkdirAll(storeDir, 0755); err != nil {
		return "", fmt.Errorf("Could not access store file at %s, %s", storeDir, err)
	}

	return storeDir, nil
}
