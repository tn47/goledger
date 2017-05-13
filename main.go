package main

import "os"
import "fmt"
import "flag"

import "github.com/prataprc/golog"
import "github.com/tn47/goledger/dblentry"
import "github.com/tn47/goledger/api"

var options struct {
	dbname   string
	journals []string
	loglevel string
}

func argparse() []string {
	var journals string

	f := flag.NewFlagSet("ledger", flag.ExitOnError)
	f.Usage = func() {
		fmsg := "Usage of command: %v [ARGS]\n"
		fmt.Printf(fmsg, os.Args[0])
		f.PrintDefaults()
	}

	f.StringVar(&journals, "f", "example/first.ldg",
		"comma separated list of input files")
	f.StringVar(&options.dbname, "db", "devjournal",
		"provide datastore name")
	f.StringVar(&options.loglevel, "log", "warn",
		"console log level")
	f.Parse(os.Args[1:])

	cwd, err := os.Getwd()
	if err != nil {
		fmt.Printf("os.Getwd(): %v\n", err)
		os.Exit(1)
	}

	if journals == "list" {
		options.journals, err = listjournals(cwd)
		if err != nil {
			os.Exit(1)
		}

	} else if journals == "find" {
		options.journals, err = findjournals(cwd)
		if err != nil {
			os.Exit(1)
		}

	} else {
		options.journals = api.Parsecsv(journals)
	}
	options.journals = append(options.journals, coveringjournals(cwd)...)

	args := f.Args()

	return args
}

func main() {
	args := argparse()

	logsetts := map[string]interface{}{
		"log.level":      options.loglevel,
		"log.file":       "",
		"log.timeformat": "",
		"log.prefix":     "[%v]",
	}
	log.SetLogger(nil, logsetts)

	reporter := NewReporter(args)
	db := dblentry.NewDatastore(options.dbname, reporter)

	for _, journal := range options.journals {
		log.Debugf("processing journal %q\n", journal)
		if err := firstpass(db, journal); err != nil {
			os.Exit(1)
		}
	}
	db.Firstpassok()

	db.PrintAccounts()

	if err := secondpass(db); err != nil {
		os.Exit(2)
	}
	db.Secondpassok()

	reporter.Render(db, args)
}
