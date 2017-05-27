package main

import "os"
import "fmt"
import "flag"
import "strconv"

import "github.com/prataprc/golog"
import "github.com/tn47/goledger/api"

func argparse() ([]string, error) {
	var journals, outfile, finyear string

	f := flag.NewFlagSet("ledger", flag.ExitOnError)
	f.Usage = func() {
		fmsg := "Usage of command: %v [OPTIONS] COMMAND [ARGS]\n"
		fmt.Printf(fmsg, os.Args[0])
		f.PrintDefaults()
	}

	f.StringVar(&api.Options.Dbname, "db", "devjournal",
		"Provide datastore name")
	f.StringVar(&journals, "f", "example/first.ldg",
		"Comma separated list of input files.")
	f.StringVar(&outfile, "o", "",
		"outfile to report")
	f.StringVar(&api.Options.Currentdt, "current", "",
		"Display only transactions on or before the current date.")
	f.StringVar(&api.Options.Begindt, "begin", "",
		"Display only transactions on or before the current date.")
	f.StringVar(&api.Options.Enddt, "end", "",
		"Display only transactions on or before the current date.")
	f.StringVar(&finyear, "fy", "",
		"financial year.")
	f.StringVar(&api.Options.Period, "period", "",
		"Limit the processing to transactions in PERIOD_EXPRESSION.")
	f.BoolVar(&api.Options.Nosubtotal, "nosubtotal", false,
		"Don't accumulate postings on sub-leger to parent ledger.")
	f.BoolVar(&api.Options.Cleared, "cleared", true,
		"Display only cleared postings.")
	f.BoolVar(&api.Options.Uncleared, "uncleared", true,
		"Display only uncleared postings.")
	f.BoolVar(&api.Options.Pending, "pending", true,
		"Display only pending postings.")
	f.BoolVar(&api.Options.Onlyreal, "real", true,
		"Display only real postings.")
	f.BoolVar(&api.Options.Onlyactual, "actual", true,
		"Display only actual postings, not automated ones.")
	f.BoolVar(&api.Options.Related, "related", false,
		"Display only related postings.")
	f.BoolVar(&api.Options.Dcformat, "dc", true,
		"Display only real postings.")
	f.BoolVar(&api.Options.Strict, "strict", false,
		"Accounts, tags or commodities not previously declared "+
			"will cause warnings.")
	f.BoolVar(&api.Options.Pedantic, "pedantic", false,
		"Accounts, tags or commodities not previously declared "+
			"will cause errors.")
	f.BoolVar(&api.Options.Checkpayee, "checkpayee", false,
		"Payee not previously declared will cause error.")
	f.BoolVar(&api.Options.Verbose, "v", false,
		"verbose reporting / listing")

	f.StringVar(&api.Options.Loglevel, "log", "info",
		"Console log level")
	f.Parse(os.Args[1:])

	api.Options.Journals = gatherjournals(journals)
	api.Options.Outfd = argOutfd(outfile)
	api.Options.Finyear = argFinyear(finyear)
	return f.Args(), nil
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

func argOutfd(outfile string) *os.File {
	outfd := os.Stdout
	if outfile != "" {
		fd, err := os.Create(outfile)
		if err != nil {
			log.Errorf("%v\n", err)
			os.Exit(1)
		}
		outfd = fd
	}
	return outfd
}

func argFinyear(finyear string) int {
	if finyear == "" {
		return 0
	}

	fy, err := strconv.Atoi(finyear)
	if err != nil {
		log.Errorf("arg `-fy` invalid: %v\n", err)
		os.Exit(1)
	}
	return fy
}
