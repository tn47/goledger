package main

import "time"

type DB struct {
	name    string
	entries []KV
}

type KV struct {
	k time.Time
	v interface{}
}

func NewDB(name string) *DB {
	return &DB{name: name, entries: make([]KV, 0)}
}

func (db *DB) Insert(k time.Time, v interface{}) {
	if len(db.entries) == 1 { // first entry is a corner case
		db.entries = append(db.entries, KV{k: k, v: v})
		return
	}
	for index, entry := range db.entries {
		if entry.k.After(k) {
			// create room for the new entry
			db.entries = append(db.entries, KV{k: k, v: v})
			// shift entries to the right
			i, j := len(db.entries)-2, len(db.entries)-1
			for ; i >= index; i, j = i-1, j-1 {
				db.entries[j] = db.entries[i]
			}
			// and, insert the new entry.
			db.entries[index] = KV{k: k, v: v}
			return
		}
		continue
	}
	// append to the end
	db.entries = append(db.entries, KV{k: k, v: v})
	return
}

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
