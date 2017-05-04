package main

import "sort"
import "fmt"
import "strings"

import "github.com/prataprc/goledger/dblentry"

type ReportRegister struct {
	rcf          *RCformat
	includeaccns []string
	register     map[string][]string
}

func NewReportRegister(args []string) *ReportRegister {
	report := &ReportRegister{
		rcf:      NewRCformat(),
		register: make(map[string][]string),
	}
	if len(args) > 1 {
		report.includeaccns = args[1:]
	}
	return report
}

func (report *ReportRegister) GetCallback() func(db *dblentry.Datastore,
	trans *dblentry.Transaction,
	p *dblentry.Posting,
	acc *dblentry.Account) {

	return report.callback
}

func (report *ReportRegister) Render(args []string) {
	// sort
	keys := []string{}
	for name := range report.register {
		keys = append(keys, name)
	}
	sort.Strings(keys)

	prevkey := ""
	for _, key := range keys {
		cols := report.register[key]
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

	// start printing
	fmt.Println()
	cols := []string{
		" By-date ", " Description ", " Account ", " Amount ", " Balance ",
	}
	fmt.Println(fmt.Sprintf(" %-14s%-40s%14s\n", cols[0], cols[1], cols[2]))
	for _, cols := range report.rcf.rows {
		fmt.Println(fmt.Sprintf(" %-14s%-40s%14s", cols[0], cols[1], cols[2]))
	}
	fmt.Println()
}

func (report *ReportRegister) includeaccount(accountname string) bool {
	if len(report.includeaccns) == 0 {
		return true
	}
	for _, include := range report.includeaccns {
		return strings.HasPrefix(accountname, include)
	}
	return false
}

func (report *ReportRegister) callback(
	db *dblentry.Datastore,
	trans *dblentry.Transaction,
	p *dblentry.Posting,
	acc *dblentry.Account) {

	row := []string{
		trans.Date().Format("2006-01-02"),
		trans.Description(),
		fmt.Sprintf("%s", acc.Name()),
		fmt.Sprintf("%.2f", p.Commodity().Amount()),
		fmt.Sprintf("%.2f", acc.Balance()),
	}
	report.register[acc.Name()] = row
}
