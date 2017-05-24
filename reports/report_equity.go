package reports

import "fmt"
import "time"
import "sort"

import "github.com/tn47/goledger/api"

// ReportEquity for equity reporting.
type ReportEquity struct {
	rcf            *RCformat
	filteraccounts []string
	latestdate     time.Time
	equity         map[string][][]string
}

// NewReportEquity create a new instance for equity reporting.
func NewReportEquity(args []string) *ReportEquity {
	report := &ReportEquity{
		rcf:            NewRCformat(),
		filteraccounts: make([]string, 0),
		equity:         make(map[string][][]string),
	}
	if len(args) > 1 {
		report.filteraccounts = args[1:]
	}
	return report
}

//---- api.Reporter methods

func (report *ReportEquity) Firstpass(
	db api.Datastorer, trans api.Transactor, p api.Poster) error {

	return nil
}

func (report *ReportEquity) Transaction(
	db api.Datastorer, trans api.Transactor) error {

	return nil
}

func (report *ReportEquity) Posting(
	db api.Datastorer, trans api.Transactor, p api.Poster) error {

	acc := p.Account()

	// filter account
	if api.Filterstring(acc.Name(), report.filteraccounts) == false {
		return nil
	}
	report.latestdate = trans.Date()
	// format account balance
	if balances := acc.FmtEquity(db, trans, p, acc); len(balances) > 0 {
		report.equity[acc.Name()] = balances
	} else {
		delete(report.equity, acc.Name())
	}

	return nil
}

func (report *ReportEquity) BubblePosting(
	db api.Datastorer, trans api.Transactor,
	p api.Poster, account api.Accounter) error {

	return nil
}

func (report *ReportEquity) Render(args []string, db api.Datastorer) {
	rcf := report.rcf

	// sort
	keys := []string{}
	for name := range report.equity {
		keys = append(keys, name)
	}
	sort.Strings(keys)

	cols := []string{
		report.latestdate.Format("2006/Jan/02"), "Opening Balance", "",
	}
	rcf.addrow(cols...)

	for _, key := range keys {
		rows := report.equity[key]
		for _, row := range rows {
			report.rcf.addrow(row...)
		}
	}

	w0 := rcf.maxwidth(rcf.column(0)) // Date
	w1 := rcf.maxwidth(rcf.column(1)) // Account name
	w2 := rcf.maxwidth(rcf.column(2)) // Balance (amount)
	if (w0 + w1 + w2) > 70 {
		_ /*w1*/ = rcf.FitAccountname(1, 70-w0-w2)
	}

	rcf.paddcells()
	fmsg := rcf.Fmsg(" %%-%vs%%-%vs%%%vs\n")

	// start printing
	fmt.Println()
	for _, cols := range report.rcf.rows {
		fmt.Printf(fmsg, cols[0], cols[1], cols[2])
	}
	fmt.Println()
}

func (report *ReportEquity) Clone() api.Reporter {
	nreport := *report
	nreport.rcf = report.rcf.Clone()
	nreport.filteraccounts = report.filteraccounts
	nreport.equity = make(map[string][][]string)
	return &nreport
}
