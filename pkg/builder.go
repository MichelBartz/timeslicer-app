package pkg

// StoreBuilder represents a file store builder
// it rebuilds the main store .db file based on its .index
type StoreBuilder struct {
	fs *FileStore
}

// NewStoreBuilder create a StoreBuilder and returns it
func NewStoreBuilder(fs *FileStore) *StoreBuilder {
	return &StoreBuilder{
		fs: fs,
	}
}
