package main

import "os"
import "fmt"
import "flag"

import "github.com/tn47/goledger/api"

var options struct {
	dbname     string
	journals   []string
	currentdt  string
	begindt    string
	enddt      string
	period     string
	cleared    bool
	uncleared  bool
	pending    bool
	onlyreal   bool
	onlyactual bool
	related    bool
	dcformat   bool
	strict     bool
	verbose    bool
	loglevel   string
}

func argparse() []string {
	var journals string

	f := flag.NewFlagSet("ledger", flag.ExitOnError)
	f.Usage = func() {
		fmsg := "Usage of command: %v [OPTIONS] COMMAND [ARGS]\n"
		fmt.Printf(fmsg, os.Args[0])
		f.PrintDefaults()
	}

	f.StringVar(&options.dbname, "db", "devjournal",
		"Provide datastore name")
	f.StringVar(&journals, "f", "example/first.ldg",
		"Comma separated list of input files.")
	f.StringVar(&options.currentdt, "current", "",
		"Display only transactions on or before the current date.")
	f.StringVar(&options.begindt, "begin", "",
		"Display only transactions on or before the current date.")
	f.StringVar(&options.enddt, "end", "",
		"Display only transactions on or before the current date.")
	f.StringVar(&options.period, "period", "",
		"Limit the processing to transactions in PERIOD_EXPRESSION.")
	f.BoolVar(&options.cleared, "cleared", true,
		"Display only cleared postings.")
	f.BoolVar(&options.uncleared, "uncleared", true,
		"Display only uncleared postings.")
	f.BoolVar(&options.pending, "pending", true,
		"Display only pending postings.")
	f.BoolVar(&options.onlyreal, "real", true,
		"Display only real postings.")
	f.BoolVar(&options.onlyactual, "actual", true,
		"Display only actual postings, not automated ones.")
	f.BoolVar(&options.related, "related", false,
		"Display only related postings.")
	f.BoolVar(&options.dcformat, "dc", true,
		"Display only real postings.")
	f.BoolVar(&options.strict, "strict", false,
		"Accounts, tags or commodities not previously declared "+
			"will cause warnings.")
	f.BoolVar(&options.verbose, "v", false,
		"verbose reporting / listing")

	f.StringVar(&options.loglevel, "log", "info",
		"Console log level")
	f.Parse(os.Args[1:])

	options.journals = gatherjournals(journals)

	return f.Args()
}

func gatherjournals(journals string) (files []string) {
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Printf("os.Getwd(): %v\n", err)
		os.Exit(1)
	}

	if journals == "list" {
		files, err = listjournals(cwd)
		if err != nil {
			os.Exit(1)
		}

	} else if journals == "find" {
		files, err = findjournals(cwd)
		if err != nil {
			os.Exit(1)
		}

	} else {
		files = api.Parsecsv(journals)
	}
	files = append(files, coveringjournals(cwd)...)
	return files
}
