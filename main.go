package main

import "os"
import "fmt"
import "flag"

var options struct {
	dbname   string
	journals []string
}

func argparse() []string {
	f := flag.NewFlagSet("ledger", flag.ExitOnError)
	f.Usage = func() {
		fmsg := "Usage of command: %v [ARGS]\n"
		fmt.Printf(fmsg, os.Args[0])
		f.PrintDefaults()
	}

	f.Parse(os.Args[1:])

	options.dbname = "devjournal"

	return f.Args()
}

func main() {
	options.journals = argparse()
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Printf("os.Getwd(): %v\n", err)
		os.Exit(1)
	}

	db := NewDatastore(options.dbname)

	journals := getjournals(cwd)
	journals = append(journals, options.journals...)
	for _, journal := range journals {
		firstpass(db, journal)
	}
}
