package main

import "github.com/prataprc/goledger/dblentry"

func secondpass(db *dblentry.Datastore) error {
	return db.Secondpass()
}
