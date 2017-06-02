package reports

import "strings"
import "fmt"

import "github.com/prataprc/golog"
import "github.com/tn47/goledger/api"
import "github.com/tn47/goledger/dblentry"

type ReportPassbook struct {
	rcf      *RCformat
	accname  string
	postings [][]string
}

func NewReportPassbook(args []string) (*ReportPassbook, error) {
	report := &ReportPassbook{
		rcf: NewRCformat(), postings: make([][]string, 0),
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

	acc := p.Account()
	if acc.Name() == report.accname {
		rows := acc.FmtPassbook(db, trans, p, acc)
		report.postings = append(report.postings, rows...)
	}
	return nil
}

func (report *ReportPassbook) BubblePosting(
	db api.Datastorer, trans api.Transactor,
	p api.Poster, account api.Accounter) error {

	return nil
}

func (report *ReportPassbook) Render(args []string, db api.Datastorer) {
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
