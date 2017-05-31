package reports

import "fmt"
import "reflect"

import "github.com/prataprc/golog"
import "github.com/tn47/goledger/api"

// Reports manages all reporting commands.
type Reports struct {
	reporters []api.Reporter

	// stats
	n_accounts     map[string]int64
	n_transactions int64
	n_postings     int64
}

// NewReporter create a new reporter.
func NewReporter(args []string) (reporter api.Reporter) {
	reports := &Reports{
		reporters:  make([]api.Reporter, 0),
		n_accounts: make(map[string]int64),
	}

	if len(args) == 0 {
		return reports
	}

	switch args[0] {
	case "balance", "bal":
		reports.reporters = append(reports.reporters, NewReportBalance(args))
	case "register", "reg", "r":
		reports.reporters = append(reports.reporters, NewReportRegister(args))
	case "equity":
		reports.reporters = append(reports.reporters, NewReportEquity(args))
	case "list", "ls":
		reports.reporters = append(reports.reporters, NewReportList(args))
	case "print":
		reports.reporters = append(reports.reporters, NewReportPrint(args))
	}
	return reports
}

//---- api.Reporter methods

func (reports *Reports) Firstpass(
	db api.Datastorer, trans api.Transactor, p api.Poster) error {

	reports.trystrict(db, trans, p)

	if err := reports.trypedantic(db, trans, p); err != nil {
		return err
	}

	for _, reporter := range reports.reporters {
		if err := reporter.Firstpass(db, trans, p); err != nil {
			return err
		}
	}
	return nil
}

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

	n, ok := reports.n_accounts[p.Account().Name()]
	if ok {
		n++
	} else {
		n = 0
	}
	reports.n_accounts[p.Account().Name()] = n

	reports.n_postings++

	for _, reporter := range reports.reporters {
		if err := reporter.Posting(db, trans, p); err != nil {
			return err
		}
	}

	if api.Options.Nosubtotal == false {
		if err := db.AggregateTotal(trans, p); err != nil {
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

func (reports *Reports) Render(args []string, db api.Datastorer) {
	outfd := api.Options.Outfd
	if len(args) == 0 {
		fmt.Fprintf(outfd, "  No. of transactions: %5v\n", reports.n_transactions)
		fmt.Fprintf(outfd, "  No. of postings:     %5v\n", reports.n_postings)
		fmt.Fprintf(outfd, "  No. of accounts:	%5v\n", len(reports.n_accounts))
		fmt.Fprintln(outfd)
		fmt.Fprintf(outfd, "  Accountwise postings\n")
		fmt.Fprintf(outfd, "  --------------------\n")
		for name, count := range reports.n_accounts {
			fmt.Fprintf(outfd, "  %15v %5v\n", name, count)
		}
	}

	for _, reporter := range reports.reporters {
		reporter.Render(args, db)
	}
}

func (reports *Reports) Clone() api.Reporter {
	nreports := *reports
	nreports.reporters = []api.Reporter{}
	nreports.n_accounts = map[string]int64{}
	for _, reporter := range reports.reporters {
		nreports.reporters = append(nreports.reporters, reporter.Clone())
	}
	return &nreports
}

func (reports *Reports) String() string {
	return fmt.Sprintf("Reports")
}

func (reports *Reports) trystrict(
	db api.Datastorer, trans api.Transactor, p api.Poster) {

	if api.Options.Strict == false {
		return
	}

	comm := p.Commodity()
	if comm != nil && reflect.ValueOf(comm).IsNil() == false {
		if db.IsCommodityDeclared(comm.Name()) == false {
			log.Warnf("commodity %q is not pre-declared\n", comm.Name())
		}
	}

	pr := p.Lotprice()
	if pr != nil && reflect.ValueOf(pr).IsNil() == false {
		if db.IsCommodityDeclared(pr.Name()) == false {
			log.Warnf("commodity %q not pre-declared\n", pr.Name())
		}
	}
	pr = p.Costprice()
	if pr != nil && reflect.ValueOf(pr).IsNil() == false {
		if db.IsCommodityDeclared(pr.Name()) == false {
			log.Warnf("commodity %q not pre-declared\n", pr.Name())
		}
	}
	pr = p.Balanceprice()
	if pr != nil && reflect.ValueOf(pr).IsNil() == false {
		if db.IsCommodityDeclared(pr.Name()) == false {
			log.Warnf("commodity %q not pre-declared\n", pr.Name())
		}
	}

	accname := p.Account().Name()
	if db.IsAccountDeclared(accname) == false {
		log.Warnf("account %q not pre-declared\n", accname)
	}
	if api.Options.Checkpayee {
		if payee := p.Payee(); db.IsPayeeDeclared(payee) == false {
			log.Warnf("payee %q not pre-declared\n", payee)
		}
	}
}

func (reports *Reports) trypedantic(
	db api.Datastorer, trans api.Transactor, p api.Poster) error {

	if api.Options.Pedantic == false {
		return nil
	}

	comm := p.Commodity()
	if comm != nil && reflect.ValueOf(comm).IsNil() == false {
		if db.IsCommodityDeclared(comm.Name()) == false {
			err := fmt.Errorf("commodity %q is not pre-declared", comm.Name())
			log.Errorf("%v\n", err)
			return err
		}
	}

	pr := p.Lotprice()
	if pr != nil && reflect.ValueOf(pr).IsNil() == false {
		if db.IsCommodityDeclared(pr.Name()) == false {
			err := fmt.Errorf("commodity %q not pre-declared\n", pr.Name())
			log.Errorf("%v\n", err)
			return err
		}
	}
	pr = p.Costprice()
	if pr != nil && reflect.ValueOf(pr).IsNil() == false {
		if db.IsCommodityDeclared(pr.Name()) == false {
			err := fmt.Errorf("commodity %q not pre-declared", pr.Name())
			log.Errorf("%v\n", err)
			return err
		}
	}
	pr = p.Balanceprice()
	if pr != nil && reflect.ValueOf(pr).IsNil() == false {
		if db.IsCommodityDeclared(pr.Name()) == false {
			err := fmt.Errorf("commodity %q not pre-declared", pr.Name())
			log.Errorf("%v\n", err)
			return err
		}
	}

	accname := p.Account().Name()
	if db.IsAccountDeclared(accname) == false {
		err := fmt.Errorf("account %q not declared before\n", accname)
		log.Errorf("%v\n", err)
		return err
	}
	if api.Options.Checkpayee {
		if payee := p.Payee(); db.IsPayeeDeclared(payee) == false {
			err := fmt.Errorf("payee %q not pre-declared", payee)
			log.Errorf("%v\n", err)
			return err
		}
	}
	return nil
}
