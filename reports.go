package main

import "fmt"
import "sort"

import "github.com/prataprc/goledger/dblentry"
import s "github.com/prataprc/gosettings"

type Reports struct {
	args    []string
	balance map[string][]string
	rcf     *RCformat
}

func NewReport(args []string) *Reports {
	report := &Reports{args: args}
	report.initreport(args)
	return report
}

func (report *Reports) callback(
	db *dblentry.Datastore,
	trans *dblentry.Transaction,
	p *dblentry.Posting,
	acc *dblentry.Account) {

	if len(report.args) == 0 {
		return
	}

	switch report.args[0] {
	case "balance":
		report.balancerow(trans, acc)
	}
}

func (report *Reports) initreport(args []string) {
	if len(report.args) == 0 {
		return
	}

	switch args[0] {
	case "balance":
		report.balance = make(map[string][]string)
		heads := []string{"Date", "Account", "Balance"}
		report.rcf = NewRCformat(heads, make(s.Settings))
	}
}

func (report *Reports) balancerow(
	trans *dblentry.Transaction, acc *dblentry.Account) {

	row := []string{
		trans.Date().Format("2006-01-02"),
		fmt.Sprintf("%-40s", acc.Name()),
		fmt.Sprintf("%10.2f", acc.Balance()),
	}
	report.balance[acc.Name()] = row
}

func (report *Reports) Render(args []string) {
	if len(args) == 0 {
		return
	}
	switch args[0] {
	case "balance":
		report.RenderBalance()
	}
}

func (report *Reports) RenderBalance() {
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
