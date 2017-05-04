package main

import "sort"
import "fmt"

import "github.com/prataprc/goledger/dblentry"
import s "github.com/prataprc/gosettings"

type ReportBalance struct {
	rcf     *RCformat
	balance map[string][]string
}

func NewReportBalance(args []string) *ReportBalance {
	heads := []string{"Date", "Account", "Balance"}
	report := &ReportBalance{
		rcf:     NewRCformat(heads, make(s.Settings)),
		balance: make(map[string][]string),
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
	keys := []string{}
	for name := range report.balance {
		keys = append(keys, name)
	}
	sort.Strings(keys)
	for _, key := range keys {
		report.rcf.Addrow(report.balance[key]...)
	}
	report.rcf.RenderBalance()
}

func (report *ReportBalance) callback(
	db *dblentry.Datastore,
	trans *dblentry.Transaction,
	p *dblentry.Posting,
	acc *dblentry.Account) {

	row := []string{
		trans.Date().Format("2006-01-02"),
		fmt.Sprintf("%-40s", acc.Name()),
		fmt.Sprintf("%10.2f", acc.Balance()),
	}
	report.balance[acc.Name()] = row
}
