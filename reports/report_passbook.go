package reports

import "fmt"
import "sort"
import "time"
import "strings"

import "github.com/bnclabs/golog"
import "github.com/tn47/goledger/api"
import "github.com/tn47/goledger/dblentry"

type ReportPassbook struct {
	rcf     *RCformat
	accname string
	// common to all mapreduce
	postings [][]string
	// mapreduce-2
	findates map[string]*time.Time            // payee -> time
	payees   map[string]*dblentry.DoubleEntry // payee -> de
	de       *dblentry.DoubleEntry
}

func NewReportPassbook(args []string) (*ReportPassbook, error) {
	report := &ReportPassbook{
		rcf:      NewRCformat(),
		postings: make([][]string, 0),
		findates: make(map[string]*time.Time),
		payees:   make(map[string]*dblentry.DoubleEntry),
	}

	if len(args) == 1 {
		err := fmt.Errorf("provide the passbook account")
		log.Errorf("%v\n", err)
		return nil, err
	}
	report.accname = strings.Trim(args[1], " \t")
	return report, nil
}

//---- api.Reporter method

func (report *ReportPassbook) Firstpass(
	db api.Datastorer, trans api.Transactor, p api.Poster) error {

	return nil
}

func (report *ReportPassbook) Transaction(
	db api.Datastorer, trans api.Transactor) error {

	return nil
}

func (report *ReportPassbook) Posting(
	db api.Datastorer, trans api.Transactor, p api.Poster) error {

	if api.FilterPeriod(trans.Date(), false /*nobegin*/) == false {
		return nil
	}

	if api.Options.Bypayee {
		return report.mapreduce2(db, trans, p)
	}
	return report.mapreduce1(db, trans, p)
}

func (report *ReportPassbook) BubblePosting(
	db api.Datastorer, trans api.Transactor,
	p api.Poster, account api.Accounter) error {

	return nil
}

func (report *ReportPassbook) Render(args []string, db api.Datastorer) {
	if api.Options.Bypayee {
		report.prerender2(args, db)
	}
	report.render1(args, db)
}

func (report *ReportPassbook) Clone() api.Reporter {
	nreport := *report
	nreport.rcf = report.rcf.Clone()
	nreport.postings = make([][]string, 0, len(report.postings))
	for _, posting := range report.postings {
		nreport.postings = append(nreport.postings, posting)
	}
	return &nreport
}

func (report *ReportPassbook) Startjournal(fname string, included bool) {
	panic("not implemented")
}

func (report *ReportPassbook) mapreduce1(
	db api.Datastorer, trans api.Transactor, p api.Poster) error {

	acc := p.Account()
	if acc.Name() == report.accname {
		rows := acc.FmtPassbook(db, trans, p, acc)
		report.postings = append(report.postings, rows...)
	}
	return nil
}

func (report *ReportPassbook) mapreduce2(
	db api.Datastorer, trans api.Transactor, p api.Poster) error {

	if p.Account().Name() != report.accname {
		return nil
	}

	date, payee := trans.Date(), p.Payee()
	findate, ok := report.findates[payee]
	if ok == false || findate.Before(date) {
		report.findates[payee] = &date
	}
	_, ok = report.payees[payee]
	if ok == false {
		report.payees[payee] = dblentry.NewDoubleEntry(payee)
	}
	report.payees[payee].AddBalance(p.Commodity())
	return nil
}

func (report *ReportPassbook) prerender2(args []string, db api.Datastorer) {
	payees := []string{}
	for payee := range report.payees {
		payees = append(payees, payee)
	}
	sort.Strings(payees)

	report.de = dblentry.NewDoubleEntry("passbook")
	for _, payee := range payees {
		de, balrows, balnames := report.payees[payee], [][]string{}, []string{}
		for _, bal := range de.Balances() {
			cols := []string{"", ""} // date, payee
			if bal.IsDebit() {
				cols = append(cols, bal.String(), "", "")
			} else if bal.IsCredit() {
				amtstr := bal.MakeSimilar(-bal.Amount()).String()
				cols = append(cols, "", amtstr, "")
			}
			report.de.AddBalance(bal)
			cols[len(cols)-1] = report.de.Balance(bal.Name()).String()
			balnames = append(balnames, bal.Name())
			balrows = append(balrows, cols)
		}
		for _, bal := range report.de.Balances() {
			if api.HasString(balnames, bal.Name()) {
				continue
			}
			balrows = append(balrows, []string{"", "", "", "", bal.String()})
		}
		if len(balrows) > 0 {
			balrows[0][0] = report.findates[payee].Format("2006-Jan-02")
			balrows[0][1] = payee
		}
		report.postings = append(report.postings, balrows...)
	}
}

func (report *ReportPassbook) render1(args []string, db api.Datastorer) {
	rcf := report.rcf

	cols := []string{"By-date", "Payee", "Debit", "Credit", "Balance"}
	rcf.addrow(cols...)
	rcf.addrow([]string{"", "", "", "", ""}...)

	for _, cols := range report.postings {
		report.rcf.addrow(cols...)
	}

	w0 := rcf.maxwidth(rcf.column(0)) // Date
	w1 := rcf.maxwidth(rcf.column(1)) // Payee
	w2 := rcf.maxwidth(rcf.column(2)) // Debit
	w3 := rcf.maxwidth(rcf.column(3)) // Credit
	w4 := rcf.maxwidth(rcf.column(4)) // Balance (amount)
	if (w0 + w1 + w2 + w3 + w4) > 70 {
		w1 = rcf.FitPayee(1, 70-w0-w2-w3-w4)
	}

	rcf.paddcells()
	fmsg := rcf.Fmsg(" %%-%vs%%-%vs%%-%vs%%%vs%%%vs\n")
	comm1 := dblentry.NewCommodity("")

	// start printing
	outfd := api.Options.Outfd
	fmt.Fprintln(outfd)
	for i, cols := range report.rcf.rows {
		items := []interface{}{cols[0], cols[1], cols[2], cols[3]}
		if i < 2 {
			items = append(items, cols[4])
		} else {
			items = append(items, CommodityColor(db, comm1, cols[4]))
		}
		fmt.Fprintf(outfd, fmsg, items...)
	}
	fmt.Fprintln(outfd)
}
