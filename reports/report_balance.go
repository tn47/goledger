package reports

import "strings"
import "sort"
import "time"
import "fmt"

import "github.com/prataprc/goparsec"
import "github.com/tn47/goledger/api"
import "github.com/tn47/goledger/dblentry"

// ReportBalance for balance reporting.
type ReportBalance struct {
	rcf       *RCformat
	fe        *api.Filterexpr
	balance   map[string][][]string
	de        *dblentry.DoubleEntry
	finaldate time.Time
	postings  map[string]bool
	bubbleacc map[string]bool
}

// NewReportBalance creates an instance for balance reporting
func NewReportBalance(args []string) (*ReportBalance, error) {
	report := &ReportBalance{
		rcf:       NewRCformat(),
		balance:   make(map[string][][]string),
		postings:  map[string]bool{},
		bubbleacc: map[string]bool{},
		de:        dblentry.NewDoubleEntry("finaltally"),
	}
	if len(args) > 1 {
		filterarg := api.MakeFilterexpr(args[1:])
		node, _ := api.YFilterExpr(parsec.NewScanner([]byte(filterarg)))
		if err, ok := node.(error); ok {
			return nil, err
		}
		report.fe = node.(*api.Filterexpr)
	}
	return report, nil
}

//---- api.Reporter methods

func (report *ReportBalance) Firstpass(
	db api.Datastorer, trans api.Transactor, p api.Poster) error {

	return nil
}

func (report *ReportBalance) Transaction(
	db api.Datastorer, trans api.Transactor) error {

	return nil
}

func (report *ReportBalance) Posting(
	db api.Datastorer, trans api.Transactor, p api.Poster) error {

	acc := p.Account()

	// filter account
	if report.isfiltered() && report.fe.Match(acc.Name()) == false {
		return nil
	}
	if api.FilterPeriod(trans.Date()) == false {
		return nil
	}

	// final balance
	report.de.AddBalance(p.Commodity().(*dblentry.Commodity))
	report.finaldate = trans.Date()

	// format account balance
	var balances [][]string
	if api.Options.Dcformat {
		balances = acc.FmtDCBalances(db, trans, p, acc)
	} else {
		balances = acc.FmtBalances(db, trans, p, acc)
	}

	if len(balances) > 0 {
		report.balance[acc.Name()] = balances
	} else {
		delete(report.balance, acc.Name())
	}

	report.postings[acc.Name()] = true
	return nil
}

func (report *ReportBalance) BubblePosting(
	db api.Datastorer, trans api.Transactor,
	p api.Poster, account api.Accounter) error {

	if api.Options.Nosubtotal || report.isfiltered() {
		return nil
	} else if api.FilterPeriod(trans.Date()) == false {
		return nil
	}
	bbname := account.Name()

	// format account balance
	if api.Options.Dcformat {
		report.balance[bbname] = account.FmtDCBalances(db, trans, p, account)
	} else {
		report.balance[bbname] = account.FmtBalances(db, trans, p, account)
	}

	report.bubbleacc[bbname] = true
	return nil
}

func (report *ReportBalance) Render(args []string, db api.Datastorer) {
	report.prunebubbled()

	// sort
	keys := []string{}
	for name := range report.balance {
		keys = append(keys, name)
	}
	sort.Strings(keys)

	fmtkeys := keys
	if api.Options.Nosubtotal == false {
		fmtkeys = Indent(keys)
	}

	if api.Options.Dcformat {
		report.renderDCBalance(args, keys, fmtkeys, db)
	} else {
		report.renderBalance(args, keys, fmtkeys, db)
	}
}

func (report *ReportBalance) renderBalance(
	args, keys, fmtkeys []string, db api.Datastorer) {

	rcf := report.rcf
	rcf.addrow([]string{"By-date", "Account", "Balance"}...)
	rcf.addrow([]string{"", "", ""}...) // empty line

	for i, key := range keys {
		rows := report.balance[key]
		for j, cols := range rows {
			cols[1] = ""
			if j == len(rows)-1 {
				cols[1] = fmtkeys[i]
			}
			rcf.addrow(cols...)
		}
	}

	dashes := api.Repeatstr("-", rcf.maxwidth(rcf.column(2)))
	rcf.addrow([]string{"", "", dashes}...)
	balances := report.de.Balances()
	for i, bal := range balances {
		if i < (len(balances) - 1) {
			rcf.addrow([]string{"", "", bal.String()}...)
		} else {
			date := report.finaldate.Format("2006/Jan/02")
			rcf.addrow([]string{date, "", bal.String()}...)
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
	comm := dblentry.NewCommodity("")

	// start printing
	outfd := api.Options.Outfd
	fmt.Fprintln(outfd)
	for i, cols := range rcf.rows {
		items := []interface{}{cols[0]}
		if i < 2 {
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

func (report *ReportBalance) renderDCBalance(
	args, keys, fmtkeys []string, db api.Datastorer) {

	rcf := report.rcf

	rcf.addrow([]string{"By-date", "Account", "Debit", "Credit", "Balance"}...)
	rcf.addrow([]string{"", "", "", "", ""}...) // empty line

	for i, key := range keys {
		rows := report.balance[key]
		for j, cols := range rows {
			cols[1] = ""
			if j == len(rows)-1 {
				cols[1] = fmtkeys[i]
			}
			rcf.addrow(cols...)
		}
	}

	drdashes := api.Repeatstr("-", rcf.maxwidth(rcf.column(2)))
	crdashes := api.Repeatstr("-", rcf.maxwidth(rcf.column(3)))
	baldashes := api.Repeatstr("-", rcf.maxwidth(rcf.column(4)))
	rcf.addrow([]string{"", "", drdashes, crdashes, baldashes}...)
	balances := report.de.Balances()
	for i, bal := range balances {
		name := bal.Name()
		dr, cr := report.de.Debit(name), report.de.Credit(name)
		cols := []string{"", "", dr.String(), cr.String(), bal.String()}
		if i == (len(balances) - 1) {
			cols[0] = report.finaldate.Format("2006/Jan/02")
		}
		rcf.addrow(cols...)
	}

	w0 := rcf.maxwidth(rcf.column(0)) // Date
	w1 := rcf.maxwidth(rcf.column(1)) // Account name
	w2 := rcf.maxwidth(rcf.column(2)) // Debit (amount)
	w3 := rcf.maxwidth(rcf.column(3)) // Credit (amount)
	w4 := rcf.maxwidth(rcf.column(4)) // Balance (amount)
	if (w0 + w1 + w2 + w3 + w4) > 125 {
		_ /*w1*/ = rcf.FitAccountname(1, 125-w0-w2-w3-w4)
	}

	rcf.paddcells()
	fmsg := rcf.Fmsg(" %%-%vs%%-%vs%%%vs%%%vs%%%vs\n")
	comm1 := dblentry.NewCommodity("")
	comm2 := dblentry.NewCommodity("")
	comm3 := dblentry.NewCommodity("")

	// start printing
	outfd := api.Options.Outfd
	fmt.Fprintln(outfd)
	for i, cols := range rcf.rows {
		items := []interface{}{cols[0]}
		if i < 2 {
			items = append(items, cols[1], cols[2], cols[3], cols[4])
		} else {
			items = append(
				items,
				api.YellowFn(cols[1]),
				CommodityColor(db, comm1, cols[2]),
				CommodityColor(db, comm2, cols[3]),
				CommodityColor(db, comm3, cols[4]),
			)
		}
		fmt.Fprintf(outfd, fmsg, items...)
	}
	fmt.Fprintln(outfd)
}

func (report *ReportBalance) Clone() api.Reporter {
	nreport := *report
	nreport.rcf = report.rcf.Clone()
	nreport.fe = report.fe
	nreport.balance = make(map[string][][]string)
	nreport.de = report.de.Clone()
	nreport.postings = map[string]bool{}
	nreport.bubbleacc = map[string]bool{}
	return &nreport
}

func (report *ReportBalance) Startjournal(fname string, included bool) {
	panic("not implemented")
}

func (report *ReportBalance) prunebubbled() {
	for bbname := range report.bubbleacc {
		ln, selfpost, children := len(bbname), 0, map[string]bool{}
		for postname := range report.postings {
			if postname == bbname {
				selfpost++
			} else if strings.HasPrefix(postname, bbname) {
				parts := dblentry.SplitAccount(postname[ln+1:])
				if postname[ln] == ':' {
					children[parts[0]] = true
				}
			}
		}
		if selfpost+len(children) < 2 {
			delete(report.balance, bbname)
		}
	}
}

func (report *ReportBalance) isfiltered() bool {
	return report.fe != nil
}
