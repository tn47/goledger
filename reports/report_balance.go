package reports

import "strings"
import "sort"
import "fmt"

import "github.com/tn47/goledger/api"
import "github.com/tn47/goledger/dblentry"

// ReportBalance for balance reporting.
type ReportBalance struct {
	rcf            *RCformat
	filteraccounts []string
	balance        map[string][][]string
	finaltally     [][]string
	postings       map[string]bool
	bubbleacc      map[string]bool
}

// NewReportBalance creates an instance for balance reporting
func NewReportBalance(args []string) *ReportBalance {
	report := &ReportBalance{
		rcf:       NewRCformat(),
		balance:   make(map[string][][]string),
		postings:  map[string]bool{},
		bubbleacc: map[string]bool{},
	}
	if len(args) > 1 {
		report.filteraccounts = args[1:]
	}
	return report
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

	// final balance
	report.finaltally = db.FmtBalances(db, trans, p, acc)

	// filter account
	if api.Filterstring(acc.Name(), report.filteraccounts) == false {
		return nil
	}
	// format account balance
	if balances := acc.FmtBalances(db, trans, p, acc); len(balances) > 0 {
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

	if api.Options.Nosubtotal {
		return nil
	}

	bbname := account.Name()

	// final balance
	report.finaltally = db.FmtBalances(db, trans, p, account)

	// filter account
	if api.Filterstring(bbname, report.filteraccounts) == false {
		return nil
	}
	// format account balance
	report.balance[bbname] = account.FmtBalances(db, trans, p, account)

	report.bubbleacc[bbname] = true
	return nil
}

func (report *ReportBalance) Render(args []string, db api.Datastorer) {
	report.prunebubbled()

	rcf := report.rcf
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

	if report.isfiltered() == false {
		dashes := api.Repeatstr("-", rcf.maxwidth(rcf.column(2)))
		rcf.addrow([]string{"", "", dashes}...)
		for _, row := range report.finaltally {
			rcf.addrow(row...)
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

func (report *ReportBalance) Clone() api.Reporter {
	nreport := *report
	nreport.rcf = report.rcf.Clone()
	nreport.filteraccounts = report.filteraccounts
	nreport.balance = make(map[string][][]string)
	nreport.finaltally = [][]string{}
	nreport.postings = map[string]bool{}
	nreport.bubbleacc = map[string]bool{}
	return &nreport
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
	return len(report.filteraccounts) > 0
}
