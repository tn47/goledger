package main

import "sort"
import "fmt"
import "strings"

import "github.com/prataprc/goledger/dblentry"

type ReportBalance struct {
	rcf          *RCformat
	includeaccns []string
	balance      map[string][]string
	finaltally   []string
}

func NewReportBalance(args []string) *ReportBalance {
	report := &ReportBalance{
		rcf:     NewRCformat(),
		balance: make(map[string][]string),
	}
	if len(args) > 1 {
		report.includeaccns = args[1:]
	}
	return report
}

func (report *ReportBalance) GetCallback() func(db *dblentry.Datastore,
	trans *dblentry.Transaction,
	p *dblentry.Posting,
	acc *dblentry.Account) {

	return report.callback
}

func (report *ReportBalance) Render(args []string) {
	rcf := report.rcf

	// sort
	keys := []string{}
	for name := range report.balance {
		keys = append(keys, name)
	}
	sort.Strings(keys)

	prevkey := ""
	for _, key := range keys {
		cols := report.balance[key]
		if report.includeaccount(cols[1]) == false {
			continue
		}
		prefix := strings.Trim(dblentry.Lcp([]string{prevkey, key}), ":")
		if prefix != "" {
			spaces := repeatstr("  ", len(strings.Split(prefix, ":")))
			cols[1] = spaces + cols[1][len(prefix)+1:]
		}
		rcf.Addrow(cols...)
		prevkey = key
	}
	rcf.Addrow([]string{"", "", repeatstr("-", rcf.maxwidth(rcf.column(2)))}...)
	rcf.Addrow(report.finaltally...)

	w1 := rcf.maxwidth(rcf.column(0)) // date
	w2 := rcf.maxwidth(rcf.column(1)) // account name
	w3 := rcf.maxwidth(rcf.column(2)) // balance (amount)
	if (w1 + w2 + w3) > 70 {
		w2 = rcf.FitAccountname(1, 70-w1-w3)
	}

	rcf.Paddcells()
	fmsg := rcf.Fmsg(" %%-%vs%%-%vs%%%vs\n")

	// start printing
	fmt.Println()
	cols := []string{" By-date ", " Account ", " Balance "}
	fmt.Printf(fmsg, cols[0], cols[1], cols[2])
	fmt.Println()
	for _, cols := range rcf.rows {
		fmt.Printf(fmsg, cols[0], cols[1], cols[2])
	}
	fmt.Println()
}

func (report *ReportBalance) callback(
	db *dblentry.Datastore,
	trans *dblentry.Transaction,
	p *dblentry.Posting,
	acc *dblentry.Account) {

	row := []string{
		report.latestdate(acc.Name(), trans.Date().Format("2006/01/02")),
		fmt.Sprintf("%s", acc.Name()),
		fmt.Sprintf("%10.2f", acc.Balance()),
	}

	report.balance[acc.Name()] = row
	report.finaltally = []string{
		report.latestdate("_fulltally_", trans.Date().Format("2006/01/02")),
		"",
		fmt.Sprintf("%10.2f", db.Balance()),
	}
}

func (report *ReportBalance) latestdate(accname, date string) string {
	if cols, ok := report.balance[accname]; ok {
		if cols[0] > date {
			return cols[0]
		}
	}
	return date
}

func (report *ReportBalance) includeaccount(accountname string) bool {
	if len(report.includeaccns) == 0 {
		return true
	}
	for _, include := range report.includeaccns {
		return strings.HasPrefix(accountname, include)
	}
	return false
}
