package main

import "os"
import "fmt"
import "flag"
import "time"
import "strconv"

import "github.com/prataprc/goparsec"
import "github.com/prataprc/golog"
import "github.com/tn47/goledger/api"
import "github.com/tn47/goledger/dblentry"

func argparse() ([]string, error) {
	var journals, outfile, finyear, begindt, enddt string

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
	f.StringVar(&begindt, "begin", "",
		"Display only transactions on or before the current date.")
	f.StringVar(&enddt, "end", "",
		"Display only transactions on or before the current date.")
	f.StringVar(&finyear, "fy", "",
		"financial year.")
	f.StringVar(&api.Options.Period, "period", "",
		"Limit the processing to transactions in PERIOD_EXPRESSION.")
	f.BoolVar(&api.Options.Nosubtotal, "nosubtotal", false,
		"Don't accumulate postings on sub-leger to parent ledger.")
	f.BoolVar(&api.Options.Subtotal, "subtotal", false,
		"all transactions to be collapsed into a single, transaction")
	f.BoolVar(&api.Options.Cleared, "cleared", true,
		"Display only cleared postings.")
	f.BoolVar(&api.Options.Uncleared, "uncleared", true,
		"Display only uncleared postings.")
	f.BoolVar(&api.Options.Pending, "pending", true,
		"Display only pending postings.")
	f.BoolVar(&api.Options.Dcformat, "dc", false,
		"Display only real postings.")
	f.BoolVar(&api.Options.Strict, "strict", false,
		"Accounts, tags or commodities not previously declared "+
			"will cause warnings.")
	f.BoolVar(&api.Options.Pedantic, "pedantic", false,
		"Accounts, tags or commodities not previously declared "+
			"will cause errors.")
	f.BoolVar(&api.Options.Checkpayee, "checkpayee", false,
		"Payee not previously declared will cause error.")
	f.BoolVar(&api.Options.Stitch, "stitch", false,
		"Skip payees with `Opening Balance`")
	f.BoolVar(&api.Options.Nopl, "nopl", false,
		"skip income and expense accounts")
	f.BoolVar(&api.Options.Onlypl, "onlypl", false,
		"skip accounts other than income and expense")
	f.BoolVar(&api.Options.Detailed, "detailed", false,
		"for register, passbook commands list details")
	f.BoolVar(&api.Options.Bypayee, "bypayee", false,
		"Group postings by common payee names")
	f.BoolVar(&api.Options.Daily, "daily", false,
		"Group postings by day")
	f.BoolVar(&api.Options.Weekly, "weekly", false,
		"Group postings by week")
	f.BoolVar(&api.Options.Monthly, "monthly", false,
		"Group postings by month")
	f.BoolVar(&api.Options.Quarterly, "quarterly", false,
		"Group postings by quarter")
	f.BoolVar(&api.Options.Yearly, "yearly", false,
		"Group postings by yearly")
	f.BoolVar(&api.Options.Verbose, "v", false,
		"verbose reporting / listing")

	f.StringVar(&api.Options.Loglevel, "log", "info",
		"Console log level")
	f.Parse(os.Args[1:])

	logsetts := map[string]interface{}{
		"log.level":      api.Options.Loglevel,
		"log.file":       "",
		"log.timeformat": "",
		"log.prefix":     "%v:",
		"log.colorfatal": "red",
		"log.colorerror": "hired",
		"log.colorwarn":  "yellow",
	}
	log.SetLogger(nil, logsetts)

	api.Options.Journals = gatherjournals(journals)
	api.Options.Outfd = argOutfd(outfile)

	endyear := argFinyear(finyear)
	if endyear > 0 {
		till := time.Date(endyear, 4, 1, 0, 0, 0, 0, time.Local)
		ok := api.ValidateDate(till, endyear, 4, 1, 0, 0, 0)
		if ok == false {
			err := fmt.Errorf("invalid finyear %v", endyear)
			log.Errorf("%v\n", err)
			return nil, err
		}
		from := time.Date(endyear-1, 4, 1, 0, 0, 0, 0, time.Local)
		ok = api.ValidateDate(from, endyear-1, 4, 1, 0, 0, 0)
		if ok == false {
			err := fmt.Errorf("invalid begin year for finyear %v", endyear)
			log.Errorf("%v\n", err)
			return nil, err
		}
		// Begindt is inclusive, but not Tilldt
		api.Options.Begindt, api.Options.Enddt = &from, &till
	}

	if begindt != "" {
		scanner := parsec.NewScanner([]byte(begindt))
		node, _ := dblentry.Ydate(time.Now().Year())(scanner)
		tm, ok := node.(time.Time)
		if ok == false {
			err := fmt.Errorf("invalid date %q: %v\n", begindt, node)
			log.Errorf("%v\n", err)
			return nil, err
		}
		api.Options.Begindt = &tm
	}

	if enddt != "" {
		scanner := parsec.NewScanner([]byte(enddt))
		node, _ := dblentry.Ydate(time.Now().Year())(scanner)
		tm, ok := node.(time.Time)
		if ok == false {
			err := fmt.Errorf("invalid date %q: %v\n", enddt, node)
			log.Errorf("%v\n", err)
			return nil, err
		}
		api.Options.Enddt = &tm
	}
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
		log.Errorf("arg `-fy` %v\n", err)
		os.Exit(1)
	}
	return fy
}
