package main

import "os"

import "github.com/prataprc/golog"
import "github.com/tn47/goledger/dblentry"
import "github.com/tn47/goledger/api"

func main() {
	args := argparse()

	logsetts := map[string]interface{}{
		"log.level":      options.loglevel,
		"log.file":       "",
		"log.timeformat": "",
		"log.prefix":     "[%v]",
	}
	log.SetLogger(nil, logsetts)

	if trycommand(args) {
		os.Exit(0)
	}

	reporter := NewReporter(args)
	db := dblentry.NewDatastore(options.dbname, reporter)

	// firstpass
	for _, journal := range options.journals {
		log.Debugf("processing journal %q\n", journal)
		if err := dofirstpass(db, journal); err != nil {
			os.Exit(1)
		}
	}
	db.Firstpassok()
	db.PrintAccounts()

	// secondpass
	nreporter := reporter.Clone()
	//nreporter.secondpass()
	ndb := db.Clone(nreporter)
	if err := secondpass(ndb); err != nil {
		os.Exit(2)
	}
	ndb.Secondpassok()
	nreporter.Render(ndb, args)
}

func trycommand(args []string) bool {
	if len(args) == 0 {
		return false
	}
	switch args[0] {
	case "version", "ver":
		log.Consolef("goledger version - goledger%v\n", api.LedgerVersion)
		return true
	}
	return false
}
