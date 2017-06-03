package main

import "os"
import "fmt"

import "github.com/prataprc/golog"
import "github.com/tn47/goledger/dblentry"
import "github.com/tn47/goledger/reports"
import "github.com/tn47/goledger/api"

var _ = fmt.Sprintf("dummy")

func main() {
	args := phase1()
	reporter, db := phase2(args)
	nreporter, ndb := phase3(args, reporter, db)
	nreporter.Render(args, ndb)
}

func trycommand(args []string, phase string) bool {
	if len(args) == 0 {
		return false
	}
	switch phase {
	case "phase1":
		switch args[0] {
		case "version", "ver":
			log.Consolef("goledger version - goledger%v\n", api.LedgerVersion)
			return true
		}

	case "phase2":
	case "phase3":
	}
	return false
}

// 1. arguments are parsed.
// 2. log is initialized.
func phase1() (args []string) {
	defer func() {
		if trycommand(args, "phase1") {
			os.Exit(0)
		}
	}()

	var err error

	if args, err = argparse(); err != nil {
		os.Exit(1)
	}
	return args
}

// 1. create one or more reporter
// 2. create a datastore.
// 3. firstpass on all journal files on the datastore.
// 4. firstpass completed
func phase2(args []string) (api.Reporter, api.Datastorer) {
	defer func() {
		if trycommand(args, "phase2") {
			os.Exit(0)
		}
	}()

	reporter, err := reports.NewReporter(args)
	if err != nil {
		os.Exit(1)
	}
	db := dblentry.NewDatastore(api.Options.Dbname, reporter)

	// apply command line arguments here.
	if api.Options.Enddt != nil {
		db.Applytill(*api.Options.Enddt)
	}

	for _, journal := range api.Options.Journals {
		log.Debugf("processing journal %q\n", journal)
		reporter.Startjournal(journal, false /*included*/)
		if err := dofirstpass(reporter, db, journal); err != nil {
			os.Exit(1)
		}
	}
	db.Firstpassok()
	db.PrintAccounts() // for debug
	return reporter, db
}

// 1. clone reporter and datastore for secondpass.
// 2. do secondpass on datastore.
// 3. secondpass completed.
func phase3(
	args []string,
	reporter api.Reporter, db api.Datastorer) (api.Reporter, api.Datastorer) {

	defer func() {
		if trycommand(args, "phase3") {
			os.Exit(0)
		}
	}()

	if len(args) == 0 {
		return reporter, nil
	}

	switch args[0] {
	case "list", "ls":
		return reporter, db
	}

	nreporter := reporter.Clone()
	//nreporter.secondpass()
	ndb := db.Clone(nreporter)
	if err := secondpass(ndb); err != nil {
		os.Exit(2)
	}
	ndb.Secondpassok()
	return nreporter, ndb
}
