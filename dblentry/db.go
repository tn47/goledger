package dblentry

import "time"
import "sort"

// DB impelements array based sorted key-value entries.
type DB struct {
	name    string
	entries []KV
}

// KV entry within DB.
type KV struct {
	k time.Time
	v interface{}
}

// NewDB create a new DB.
func NewDB(name string) *DB {
	return &DB{name: name, entries: make([]KV, 0)}
}

// Insert a KV entry into DB.
func (db *DB) Insert(k time.Time, v interface{}) error {
	db.entries = append(db.entries, KV{k: k, v: v})
	sort.Sort(db)
	return nil
}

// Range over DB entries, from low to high with specified include option.
func (db *DB) Range(
	low *time.Time, high *time.Time, incl string, entries []KV) []KV {

	var entry KV

	index := 0

	// find start
	if low != nil {
		lowk := *low
	startloop:
		for index, entry = range db.entries {
			switch incl {
			case "low", "both":
				if !entry.k.Before(lowk) {
					break startloop
				}
			default:
				if entry.k.After(lowk) {
					break startloop
				}
			}
		}
	}

	var highk time.Time
	if high != nil {
		highk = *high
	}

endloop:
	for _, entry = range db.entries[index:] {
		if high != nil {
			switch incl {
			case "high", "both":
				if entry.k.After(highk) {
					break endloop
				}
			default:
				if !entry.k.Before(highk) {
					break endloop
				}
			}
		}
		entries = append(entries, entry)
	}
	return entries
}

//---- sort.Interface{} methods.

func (db *DB) Len() int {
	return len(db.entries)
}

func (db *DB) Less(i, j int) bool {
	return db.entries[i].k.Before(db.entries[j].k)
}

func (db *DB) Swap(i, j int) {
	db.entries[i], db.entries[j] = db.entries[j], db.entries[i]
}
