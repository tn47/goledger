package main

import "github.com/prataprc/goledger/dblentry"
import "github.com/prataprc/golog"

func secondpass(db *dblentry.Datastore) error {
	log.Debugf("secondpass\n")
	if err := db.Secondpass(); err != nil {
		log.Errorf("%v\n", err)
		return err
	}
	return nil
}
