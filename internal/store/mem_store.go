package store

import (
	"github.com/tidwall/buntdb"
	"log"
	"strings"
)

// DB representation of a Store.
// Only one table/bucket which contains the urls, so it's not a fully Database,
// it works only with single bucket because that all we need.
type DB struct {
	db *buntdb.DB
}

var localStore Store = &DB{}

// openDatabase open a new database connection
//
// and returns its instance.
func (d *DB) openDatabase() *buntdb.DB {
	db, err := buntdb.Open(":memory:") // Open a file that does not persist to disk.
	if err != nil {
		log.Fatal(err)
	}
	return db
}

// Set sets a shorten url and its key
//
// Note: Caller is responsible to generate a key.
func (d *DB) Set(key string, value string) error {
	return d.db.Update(func(tx *buntdb.Tx) error {
		_, _, err := tx.Set(key, value, nil)
		return err
	})
}

// Clear clears all the database entries for the table urls.
func (d *DB) Clear() error {
	return d.db.Update(func(tx *buntdb.Tx) error {
		return tx.DeleteAll()
	})
}

// Get returns a url by its key.
//
// Returns an empty string if not found.
func (d *DB) GetAll() map[string]string {
	values := make(map[string]string)
	_ = d.db.View(func(tx *buntdb.Tx) error {
		err := tx.Ascend("", func(key, value string) bool {
			values[key] = value
			return true
		})
		return err
	})

	return values
}

// Get returns a url by its key.
//
// Returns an empty string if not found.
func (d *DB) Get(key string) (value string) {
	_ = d.db.Update(func(tx *buntdb.Tx) error {
		val, err := tx.Get(key)
		if err != nil {
			return err
		}
		value = val
		return nil
	})
	return
}

// GetByValue returns all keys for a specific (original) url value.
func (d *DB) GetByValue(search string) (keys []string) {
	_ = d.db.View(func(tx *buntdb.Tx) error {
		err := tx.Ascend("", func(key, value string) bool {
			if strings.Compare(search, value) == 0 {
				keys = append(keys, key)
			}
			return true
		})
		return err
	})

	return
}

// Len returns all the "shorted" urls length
func (d *DB) Len() (num int) {
	_ = d.db.View(func(tx *buntdb.Tx) error {
		// Assume bucket exists and has keys
		num, _ = tx.Len()
		return nil
	})
	return
}

// Close shutdowns the data(base) connection.
func (d *DB) Close() {
	if err := d.db.Close(); err != nil {
		log.Fatal(err)
	}
}

// NewDB returns a new DB instance, its connection is opened.
//
// DB implements the Store.
func NewDB() *DB {
	return &DB{
		db: localStore.openDatabase(),
	}
}
