package reports

import "fmt"
import "time"
import "sort"
import "strings"

import "github.com/prataprc/goparsec"
import "github.com/tn47/goledger/api"
import "github.com/tn47/goledger/dblentry"

// ReportEquity for equity reporting.
type ReportEquity struct {
	rcf        *RCformat
	fe         *api.Filterexpr
	latestdate time.Time
	equity     map[string][][]string
}

// NewReportEquity create a new instance for equity reporting.
func NewReportEquity(args []string) (*ReportEquity, error) {
	report := &ReportEquity{
		rcf:    NewRCformat(),
		equity: make(map[string][][]string),
	}
	api.Options.Nosubtotal = true
	if len(args) > 1 {
		filterarg := api.MakeFilterexpr(args[1:])
		node, _ := api.YExpr(parsec.NewScanner([]byte(filterarg)))
		if err, ok := node.(error); ok {
			return nil, err
		}
		report.fe = node.(*api.Filterexpr)
	}
	return report, nil
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
	if report.isfiltered() && report.fe.Match(acc.Name()) == false {
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
	fmsg := rcf.Fmsg("%%-%vs%%-%vs%%%vs\n")
	comm := dblentry.NewCommodity("")

	report.rcf.rows[0][0] = strings.TrimLeft(report.rcf.rows[0][0], " ")
	// start printing
	outfd := api.Options.Outfd
	fmt.Fprintln(outfd)
	for i, cols := range report.rcf.rows {
		items := []interface{}{cols[0]}
		if i < 1 {
			items = append(items, cols[1], cols[2])
		} else {
			items = append(
				items,
				api.YellowFn(cols[1]),
				CommodityColor(db, comm, cols[2]),
			)
		}
		fmt.Fprintf(outfd, fmsg, items...)
	}
	fmt.Fprintln(outfd)
}

func (report *ReportEquity) Clone() api.Reporter {
	nreport := *report
	nreport.rcf = report.rcf.Clone()
	nreport.fe = report.fe
	nreport.equity = make(map[string][][]string)
	return &nreport
}

func (report *ReportEquity) Startjournal(fname string, included bool) {
	panic("not implemented")
}

func (report *ReportEquity) isfiltered() bool {
	return report.fe != nil
}
