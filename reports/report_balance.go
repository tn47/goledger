package reports

import "sort"
import "fmt"
import "strings"

import "github.com/prataprc/goledger/api"
import "github.com/prataprc/goledger/dblentry"

type ReportBalance struct {
	rcf            *RCformat
	filteraccounts []string
	balance        map[string][]string
	finaltally     []string
}

func NewReportBalance(args []string) *ReportBalance {
	report := &ReportBalance{
		rcf:     NewRCformat(),
		balance: make(map[string][]string),
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
		cols := report.balance[key]
		if api.Filterstring(cols[1], report.filteraccounts) == false {
			continue
		}
		prefix := strings.Trim(dblentry.AccountLcp([]string{prevkey, key}), ":")
		if prefix != "" {
			spaces := api.Repeatstr("  ", len(strings.Split(prefix, ":")))
			cols[1] = spaces + cols[1][len(prefix)+1:]
		}
		rcf.Addrow(cols...)
		prevkey = key
	}
	if report.isfiltered() == false {
		dashes := api.Repeatstr("-", rcf.maxwidth(rcf.column(2)))
		rcf.Addrow([]string{"", "", dashes}...)
		rcf.Addrow(report.finaltally...)
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

	row := []string{
		report.latestdate(acc.Name(), trans.Date().Format("2006/Jan/02")),
		fmt.Sprintf("%s", acc.Name()),
		fmt.Sprintf("%s", BalanceRepr(acc.Balances())),
	}

	report.balance[acc.Name()] = row
	report.finaltally = []string{
		report.latestdate("_fulltally_", trans.Date().Format("2006/Jan/02")),
		"",
		fmt.Sprintf("%s", BalanceRepr(db.Balances())),
	}
	return nil
}

func (report *ReportBalance) latestdate(accname, date string) string {
	if cols, ok := report.balance[accname]; ok {
		if cols[0] > date {
			return cols[0]
		}
	}
	return date
}

func (report *ReportBalance) isfiltered() bool {
	return len(report.filteraccounts) > 0
}
