package main

import "sort"
import "fmt"
import "strings"

import "github.com/prataprc/goledger/dblentry"

type ReportBalance struct {
	rcf          *RCformat
	includeaccns []string
	balance      map[string][]string
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
			spaces := ""
			for i := 0; i < len(strings.Split(prefix, ":")); i++ {
				spaces += "  "
			}
			cols[1] = spaces + cols[1][len(prefix)+1:]
		}
		report.rcf.Addrow(cols...)
		prevkey = key
	}
	report.rcf.FitWidth([]int{14, 40, 14})

	// start printing
	fmt.Println()
	cols := []string{" By-date ", " Account ", " Balance "}
	fmt.Println(fmt.Sprintf(" %-14s%-40s%14s\n", cols[0], cols[1], cols[2]))
	for _, cols := range report.rcf.rows {
		fmt.Println(fmt.Sprintf(" %-14s%-40s%14s", cols[0], cols[1], cols[2]))
	}
	fmt.Println()
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

func (report *ReportBalance) callback(
	db *dblentry.Datastore,
	trans *dblentry.Transaction,
	p *dblentry.Posting,
	acc *dblentry.Account) {

	row := []string{
		trans.Date().Format("2006-01-02"),
		fmt.Sprintf("%s", acc.Name()),
		fmt.Sprintf("%10.2f", acc.Balance()),
	}
	report.balance[acc.Name()] = row
}
