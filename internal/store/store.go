package store

import "github.com/tidwall/buntdb"

// Store is the store interface for urls.
// Note: no Del functionality.
type Store interface {
	openDatabase() *buntdb.DB
	Set(key string, value string) error // error if something went wrong
	Get(key string) string              // empty value if not found
	GetAll() map[string]string          // empty value if not found
	Len() int                           // should return the number of all the records/tables/buckets
	Close()                             // release the store or ignore
}
