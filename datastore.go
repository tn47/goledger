package main

type Datastore struct {
	name string

	// working fields
	year       int
	month      int
	dateformat string
}

func NewDatastore(name string) *Datastore {
	db := &Datastore{
		name: name,
		// working fields
		year:       -1,
		month:      -1,
		dateformat: "%d-%m-%y",
	}
	return db
}

func (db *Datastore) Year() int {
	return db.year
}

func (db *Datastore) Month() int {
	return db.month
}

func (db *Datastore) Dateformat() string {
	return db.dateformat
}
