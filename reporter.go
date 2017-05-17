package main

import "fmt"

import "github.com/tn47/goledger/api"

// Reports manages all reporting commands.
type Reports struct {
	reporters []api.Reporter
	accounts  map[string]int64

	// stats
	n_transactions int64
	n_postings     int64
}

// NewReporter create a new reporter.
func NewReporter(args []string) (reporter api.Reporter) {
	reports := &Reports{
		reporters: make([]api.Reporter, 0),
		accounts:  make(map[string]int64),
	}

	if len(args) == 0 {
		return reports
	}

	switch args[0] {
	case "balance", "bal":
		reports.reporters = append(reports.reporters, NewReportBalance(args))
	case "register", "reg":
		reports.reporters = append(reports.reporters, NewReportRegister(args))
	case "equity":
		reports.reporters = append(reports.reporters, NewReportEquity(args))
	}
	return reports
}

//---- api.Reporter methods

func (reports *Reports) Transaction(
	db api.Datastorer, trans api.Transactor) error {

	reports.n_transactions++
	for _, reporter := range reports.reporters {
		if err := reporter.Transaction(db, trans); err != nil {
			return err
		}
	}
	return nil
}

func (reports *Reports) Posting(
	db api.Datastorer, trans api.Transactor, p api.Poster) error {

	n, ok := reports.accounts[p.Account().Name()]
	if ok {
		n++
	} else {
		n = 0
	}
	reports.accounts[p.Account().Name()] = n

	reports.n_postings++

	for _, reporter := range reports.reporters {
		if err := reporter.Posting(db, trans, p); err != nil {
			return err
		}
	}
	return nil
}

func (reports *Reports) BubblePosting(
	db api.Datastorer, trans api.Transactor,
	p api.Poster, account api.Accounter) error {

	for _, reporter := range reports.reporters {
		if err := reporter.BubblePosting(db, trans, p, account); err != nil {
			return err
		}
	}
	return nil
}

func (reports *Reports) Render(db api.Datastorer, args []string) {
	if len(args) == 0 {
		fmt.Printf("  No. of transactions: %5v\n", reports.n_transactions)
		fmt.Printf("  No. of postings:     %5v\n", reports.n_postings)
		fmt.Printf("  No. of accounts:	%5v\n", len(reports.accounts))
		fmt.Println()
		fmt.Printf("  Accountwise postings\n")
		fmt.Printf("  --------------------\n")
		for name, count := range reports.accounts {
			fmt.Printf("  %15v %5v\n", name, count)
		}
	}

	for _, reporter := range reports.reporters {
		reporter.Render(db, args)
	}
}

func (reports *Reports) Clone() api.Reporter {
	nreports := *reports
	nreports.reporters = []api.Reporter{}
	nreports.accounts = map[string]int64{}
	for _, reporter := range reports.reporters {
		nreports.reporters = append(nreports.reporters, reporter.Clone())
	}
	return &nreports
}

func (reports *Reports) String() string {
	return fmt.Sprintf("Reports")
}
