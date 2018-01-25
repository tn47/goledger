package reports

import "fmt"
import "time"
import "reflect"

import "github.com/bnclabs/golog"
import "github.com/tn47/goledger/api"

// Reports manages all reporting commands.
type Reports struct {
	reporters []api.Reporter

	// stats
	journals       []string
	startdate      *time.Time
	enddate        *time.Time
	n_transactions map[uint64]bool
	n_accounts     map[string]int64
	n_commodities  map[string]int64
	n_payees       map[string]int64
	n_postings     int64
}

// NewReporter create a new reporter.
func NewReporter(args []string) (reporter api.Reporter, err error) {
	reports := &Reports{
		reporters:      make([]api.Reporter, 0),
		journals:       []string{},
		n_transactions: make(map[uint64]bool),
		n_accounts:     make(map[string]int64),
		n_commodities:  make(map[string]int64),
		n_payees:       make(map[string]int64),
	}

	if len(args) == 0 {
		return reports, nil
	}

	switch args[0] {
	case "balance", "bal", "b":
		reporter, err = NewReportBalance(args)
		reports.reporters = append(reports.reporters, reporter)
	case "register", "reg", "r":
		reporter, err = NewReportRegister(args)
		reports.reporters = append(reports.reporters, reporter)
	case "equity", "eq":
		reporter, err = NewReportEquity(args)
		reports.reporters = append(reports.reporters, reporter)
	case "list", "ls":
		reports.reporters = append(reports.reporters, NewReportList(args))
	case "print", "p":
		reports.reporters = append(reports.reporters, NewReportPrint(args))
	case "passbook", "pb", "pbook":
		reporter, err = NewReportPassbook(args)
		reports.reporters = append(reports.reporters, reporter)
	default:
		log.Errorf("invalid command %q\n", args[0])
	}
	return reports, err
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

	date := trans.Date()
	if reports.startdate == nil {
		reports.startdate = &date
	} else {
		reports.enddate = &date
	}
	reports.n_transactions[trans.Crc64()] = true

	n, ok := reports.n_accounts[p.Account().Name()]
	if ok {
		n++
	}
	reports.n_accounts[p.Account().Name()] = n
	n, ok = reports.n_commodities[p.Commodity().Name()]
	if ok {
		n++
	}
	reports.n_commodities[p.Commodity().Name()] = n
	n, ok = reports.n_payees[p.Payee()]
	if ok {
		n++
	}
	reports.n_payees[p.Payee()] = n
	reports.n_postings++

	return nil
}

func (reports *Reports) Transaction(
	db api.Datastorer, trans api.Transactor) error {

	for _, reporter := range reports.reporters {
		if err := reporter.Transaction(db, trans); err != nil {
			return err
		}
	}
	return nil
}

func (reports *Reports) Posting(
	db api.Datastorer, trans api.Transactor, p api.Poster) error {

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
		for _, s := range reports.journals {
			fmt.Fprintf(outfd, "%v\n", s)
		}

		n_postings := reports.n_postings
		if reports.startdate != nil {
			startdt := reports.startdate.Format("2006/Jan/02")
			enddt := reports.enddate.Format("2006/Jan/02")
			fmt.Fprintf(outfd, "transactions from %q to %q\n", startdt, enddt)
		}

		fmsg := "%v postings in %v transactions\n"
		fmt.Fprintf(outfd, fmsg, n_postings, len(reports.n_transactions))
		fmsg = "%v postings to %v accounts\n"
		fmt.Fprintf(outfd, fmsg, n_postings, len(reports.n_accounts))
		fmsg = "%v postings using %v commodity\n"
		fmt.Fprintf(outfd, fmsg, n_postings, len(reports.n_commodities))
		fmsg = "%v postings with %v payees\n"
		fmt.Fprintf(outfd, fmsg, n_postings, len(reports.n_payees))
		fmt.Fprintln(outfd)
	}
	if api.Options.Verbose && len(args) == 0 {
		for name, n := range reports.n_accounts {
			fmt.Fprintf(outfd, "%v postings to account %q\n", n, name)
		}
		for name, n := range reports.n_commodities {
			fmt.Fprintf(outfd, "%v postings using commodity %q\n", n, name)
		}
		for name, n := range reports.n_payees {
			fmt.Fprintf(outfd, "%v postings with payee %q\n", n, name)
		}
	}

	for _, reporter := range reports.reporters {
		reporter.Render(args, db)
	}
}

func (reports *Reports) Clone() api.Reporter {
	nreports := *reports
	nreports.reporters = []api.Reporter{}
	for _, reporter := range reports.reporters {
		nreports.reporters = append(nreports.reporters, reporter.Clone())
	}
	nreports.journals = []string{}
	for _, s := range reports.journals {
		nreports.journals = append(nreports.journals, s)
	}
	nreports.n_accounts = map[string]int64{}
	return &nreports
}

func (reports *Reports) Startjournal(fname string, included bool) {
	if included {
		s := fmt.Sprintf("including journal %q ...", fname)
		reports.journals = append(reports.journals, s)
		return
	}
	s := fmt.Sprintf("processing journal %q ...", fname)
	reports.journals = append(reports.journals, s)
	return
}

func (reports *Reports) String() string {
	return fmt.Sprintf("Reports")
}

func (reports *Reports) trystrict(
	db api.Datastorer, trans api.Transactor, p api.Poster) {

	if api.Options.Strict == false {
		return
	}

	jf := trans.Journalfile()
	comm := p.Commodity()
	if comm != nil && reflect.ValueOf(comm).IsNil() == false {
		if db.IsCommodityDeclared(comm.Name()) == false {
			fmsg := "In %q : commodity %q is not pre-declared\n"
			log.Warnf(fmsg, jf, comm.Name())
		}
	}

	pr := p.Lotprice()
	if pr != nil && reflect.ValueOf(pr).IsNil() == false {
		if db.IsCommodityDeclared(pr.Name()) == false {
			fmsg := "In %q : commodity %q is not pre-declared\n"
			log.Warnf(fmsg, jf, pr.Name())
		}
	}
	pr = p.Costprice()
	if pr != nil && reflect.ValueOf(pr).IsNil() == false {
		if db.IsCommodityDeclared(pr.Name()) == false {
			fmsg := "In %q : commodity %q is not pre-declared\n"
			log.Warnf(fmsg, jf, pr.Name())
		}
	}
	pr = p.Balanceprice()
	if pr != nil && reflect.ValueOf(pr).IsNil() == false {
		if db.IsCommodityDeclared(pr.Name()) == false {
			fmsg := "In %q : commodity %q is not pre-declared\n"
			log.Warnf(fmsg, jf, pr.Name())
		}
	}

	accname := p.Account().Name()
	if db.IsAccountDeclared(accname) == false {
		log.Warnf("In %q : account %q not pre-declared\n", jf, accname)
	}
	if api.Options.Checkpayee {
		if payee := p.Payee(); db.IsPayeeDeclared(payee) == false {
			log.Warnf("In %q : payee %q not pre-declared\n", jf, payee)
		}
	}
}

func (reports *Reports) trypedantic(
	db api.Datastorer, trans api.Transactor, p api.Poster) error {

	if api.Options.Pedantic == false {
		return nil
	}

	jf := trans.Journalfile()
	comm := p.Commodity()
	if comm != nil && reflect.ValueOf(comm).IsNil() == false {
		if db.IsCommodityDeclared(comm.Name()) == false {
			fmsg := "In %q : commodity %q is not pre-declared"
			err := fmt.Errorf(fmsg, jf, comm.Name())
			log.Errorf("%v\n", err)
			return err
		}
	}

	pr := p.Lotprice()
	if pr != nil && reflect.ValueOf(pr).IsNil() == false {
		if db.IsCommodityDeclared(pr.Name()) == false {
			fmsg := "In %q : commodity %q not pre-declared\n"
			err := fmt.Errorf(fmsg, jf, pr.Name())
			log.Errorf("%v\n", err)
			return err
		}
	}
	pr = p.Costprice()
	if pr != nil && reflect.ValueOf(pr).IsNil() == false {
		if db.IsCommodityDeclared(pr.Name()) == false {
			fmsg := "In %q : commodity %q not pre-declared"
			err := fmt.Errorf(fmsg, jf, pr.Name())
			log.Errorf("%v\n", err)
			return err
		}
	}
	pr = p.Balanceprice()
	if pr != nil && reflect.ValueOf(pr).IsNil() == false {
		if db.IsCommodityDeclared(pr.Name()) == false {
			fmsg := "In %q : commodity %q not pre-declared"
			err := fmt.Errorf(fmsg, jf, pr.Name())
			log.Errorf("%v\n", err)
			return err
		}
	}

	accname := p.Account().Name()
	if db.IsAccountDeclared(accname) == false {
		fmsg := "In %q : account %q not declared before\n"
		err := fmt.Errorf(fmsg, jf, accname)
		log.Errorf("%v\n", err)
		return err
	}
	if api.Options.Checkpayee {
		if payee := p.Payee(); db.IsPayeeDeclared(payee) == false {
			err := fmt.Errorf("In %q : payee %q not pre-declared", jf, payee)
			log.Errorf("%v\n", err)
			return err
		}
	}
	return nil
}
