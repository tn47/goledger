package main

import "os"
import "fmt"
import "flag"

import "github.com/prataprc/golog"
import "github.com/prataprc/goledger/dblentry"
import "github.com/prataprc/goledger/reports"
import "github.com/prataprc/goledger/api"

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

	options.journals = api.Parsecsv(journals)

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

	cwd, err := os.Getwd()
	if err != nil {
		fmt.Printf("os.Getwd(): %v\n", err)
		os.Exit(1)
	}

	reporter := reports.NewReporter(args)
	db := dblentry.NewDatastore(options.dbname, reporter)

	journals := getjournals(cwd)
	journals = append(journals, options.journals...)
	for _, journal := range journals {
		log.Debugf("processing journal %q\n", journal)
		if firstpass(db, journal) == false {
			os.Exit(1)
		}
	}
	if secondpass(db) == false {
		os.Exit(2)
	}
	reporter.Render(args)
}
