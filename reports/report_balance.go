package reports

import "sort"
import "fmt"
import "strings"

import "github.com/prataprc/goledger/api"
import "github.com/prataprc/goledger/dblentry"

type ReportBalance struct {
	rcf            *RCformat
	filteraccounts []string
	balance        map[string][][]string
	finaltally     [][]string
}

func NewReportBalance(args []string) *ReportBalance {
	report := &ReportBalance{
		rcf:     NewRCformat(),
		balance: make(map[string][][]string),
	}
	if len(args) > 1 {
		report.filteraccounts = args[1:]
	}
	return report
}

func (report *ReportBalance) Transaction(
	db api.Datastorer, trans api.Transactor) error {

	return nil
}

func (report *ReportBalance) Posting(
	db api.Datastorer, trans api.Transactor,
	p api.Poster, account api.Accounter) error {

	return report.posting(db, trans, p, account)
}

func (report *ReportBalance) BubblePosting(
	db api.Datastorer, trans api.Transactor,
	p api.Poster, account api.Accounter) error {

	return report.posting(db, trans, p, account)
}

func (report *ReportBalance) Render(args []string) {
	rcf := report.rcf

	// sort
	keys := []string{}
	for name := range report.balance {
		keys = append(keys, name)
	}
	sort.Strings(keys)

	rcf.Addrow([]string{"By-date", "Account", "Balance"}...)
	rcf.Addrow([]string{"", "", ""}...) // empty line

	prevkey := ""
	for _, key := range keys {
		for _, cols := range report.balance[key] {
			if cols[1] == "" {
				rcf.Addrow(cols...)
				continue
			}
			prefix := dblentry.AccountLcp([]string{prevkey, key})
			prefix = strings.Trim(prefix, ":")
			if prefix != "" {
				spaces := api.Repeatstr("  ", len(strings.Split(prefix, ":")))
				cols[1] = spaces + cols[1][len(prefix)+1:]
			}
			rcf.Addrow(cols...)
		}
		prevkey = key
	}
	if report.isfiltered() == false {
		dashes := api.Repeatstr("-", rcf.maxwidth(rcf.column(2)))
		rcf.Addrow([]string{"", "", dashes}...)
		for _, row := range report.finaltally {
			rcf.Addrow(row...)
		}
	}

	w0 := rcf.maxwidth(rcf.column(0)) // Date
	w1 := rcf.maxwidth(rcf.column(1)) // Account name
	w2 := rcf.maxwidth(rcf.column(2)) // Balance (amount)
	if (w0 + w1 + w2) > 70 {
		w1 = rcf.FitAccountname(1, 70-w0-w2)
	}

	rcf.Paddcells()
	fmsg := rcf.Fmsg(" %%-%vs%%-%vs%%%vs\n")

	// start printing
	fmt.Println()
	for _, cols := range rcf.rows {
		fmt.Printf(fmsg, cols[0], cols[1], cols[2])
	}
	fmt.Println()
}

func (report *ReportBalance) posting(
	db api.Datastorer, trans api.Transactor,
	p api.Poster, acc api.Accounter) error {

	// final balance
	report.finaltally = [][]string{}
	for _, balance := range db.Balances() {
		row := []string{"", "", balance.String()}
		report.finaltally = append(report.finaltally, row)
	}
	if ln := len(report.finaltally); ln > 0 {
		report.finaltally[ln-1][0] = trans.Date().Format("2006/Jan/02")
	}

	// filter account
	if api.Filterstring(acc.Name(), report.filteraccounts) == false {
		return nil
	}

	// format account balance
	balances, rows := acc.Balances(), [][]string{}
	var balance api.Commoditiser
	var i int
	for i, balance = range balances {
		if i == (len(balances) - 1) {
			break
		}
		row := []string{"", "", balance.String()}
		rows = append(rows, row)
	}
	finb := "0"
	if balance != nil {
		finb = balance.String()
	}
	row := []string{trans.Date().Format("2006/Jan/02"), acc.Name(), finb}
	rows = append(rows, row)

	report.balance[acc.Name()] = rows

	return nil
}

func (report *ReportBalance) isfiltered() bool {
	return len(report.filteraccounts) > 0
}
