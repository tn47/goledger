package main

import "time"
import "os"
import "fmt"
import "flag"

import "github.com/tn47/goledger/api"

var options struct {
	dbname   string
	journals []string
	fromdt   time.Time
	tilldt   time.Time
	loglevel string
}

func argparse() []string {
	var journals string

	f := flag.NewFlagSet("ledger", flag.ExitOnError)
	f.Usage = func() {
		fmsg := "Usage of command: %v [OPTIONS] COMMAND [ARGS]\n"
		fmt.Printf(fmsg, os.Args[0])
		f.PrintDefaults()
	}

	f.StringVar(&journals, "f", "example/first.ldg",
		"comma separated list of input files")
	f.StringVar(&options.dbname, "db", "devjournal",
		"provide datastore name")
	f.StringVar(&options.loglevel, "log", "info",
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

	return f.Args()
}
