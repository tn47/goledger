package main

import "fmt"

import "github.com/prataprc/golog"
import "github.com/tn47/goledger/api"

// ReportList for balance reporting.
type ReportList struct {
	rcf *RCformat
}

// NewReportList creates an instance for balance reporting
func NewReportList(args []string) *ReportList {
	report := &ReportList{rcf: NewRCformat()}
	return report
}

//---- api.Reporter methods

func (report *ReportList) Transaction(
	_ api.Datastorer, _ api.Transactor) error {

	panic("not implemented")
}

func (report *ReportList) Posting(
	_ api.Datastorer, _ api.Transactor, _ api.Poster) error {

	panic("not implemented")
}

func (report *ReportList) BubblePosting(
	_ api.Datastorer, _ api.Transactor, _ api.Poster, _ api.Accounter) error {

	panic("not implemented")
}

func (report *ReportList) Render(args []string, ndb api.Datastorer) {
	if len(args) < 2 {
		log.Errorf("insufficient arguments to list report\n")
		return
	}

	switch args[1] {
	case "accounts":
		if options.verbose == false {
			report.listAccounts(args[2:], ndb)
		} else {
			report.listAccountsV(args[2:], ndb)
		}
	}
}

func (report *ReportList) Clone() api.Reporter {
	nreport := *report
	nreport.rcf = report.rcf.Clone()
	return &nreport
}

func (report *ReportList) listAccounts(args []string, ndb api.Datastorer) {
	rcf := report.rcf
	for _, accname := range ndb.Accountnames() {
		account := ndb.GetAccount(accname)
		notes := account.Notes()
		switch len(notes) {
		case 0:
			rcf.addrow([]string{accname, ""}...)
		case 1:
			rcf.addrow([]string{accname, notes[0]}...)
		default:
			rcf.addrow([]string{accname, notes[0]}...)
			for _, note := range notes[1:] {
				rcf.addrow([]string{"", note}...)
			}
		}
	}

	rcf.paddcells()
	fmsg := rcf.Fmsg(" %%-%vs%%-%vs\n")

	// start printing
	fmt.Println()
	for _, cols := range rcf.rows {
		fmt.Printf(fmsg, cols[0], cols[1])
	}
	fmt.Println()
}

func (report *ReportList) listAccountsV(args []string, ndb api.Datastorer) {
	fmt.Println()
	for _, accname := range ndb.Accountnames() {
		account := ndb.GetAccount(accname)
		fmt.Println(account.Directive())
	}
	fmt.Println()
}
