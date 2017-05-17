package main

import "github.com/prataprc/golog"
import "github.com/tn47/goledger/api"

func secondpass(db api.Datastorer) error {
	log.Debugf("secondpass\n")
	if err := db.Secondpass(); err != nil {
		log.Errorf("%v\n", err)
		return err
	}
	return nil
}
