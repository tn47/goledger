package main

import "fmt"

import "github.com/tn47/goledger/api"

// ReportRegister for register reporting.
type ReportRegister struct {
	rcf            *RCformat
	filteraccounts []string
	filterpayees   []string
	register       [][]string
}

// NewReportRegister create an instance for register reporting.
func NewReportRegister(args []string) *ReportRegister {
	report := &ReportRegister{
		rcf:            NewRCformat(),
		filteraccounts: make([]string, 0),
		filterpayees:   make([]string, 0),
		register:       make([][]string, 0),
	}
	for i, arg := range args[1:] {
		if arg == "@" || arg == "payee" {
			report.filterpayees = append(report.filterpayees, args[i+1+1:]...)
			break
		}
		report.filteraccounts = append(report.filteraccounts, arg)
	}
	return report
}

//---- api.Reporter methods

func (report *ReportRegister) Transaction(
	db api.Datastorer, trans api.Transactor) error {

	date, transpayee := trans.Date().Format("2006-Jan-02"), trans.Payee()
	prevaccname := ""
	for _, p := range trans.GetPostings() {
		accname := p.Account().Name()
		if report.isfiltered() {
			if api.Filterstring(accname, report.filteraccounts) == false {
				continue
			}
			if api.Filterstring(p.Payee(), report.filterpayees) == false {
				continue
			}
		}
		commodity := p.Commodity()
		row := []string{
			date, transpayee, accname, commodity.String(),
			p.Account().Balance(commodity.Name()).String(),
		}
		if p.Payee() != trans.Payee() {
			row[1] = p.Payee()
		}
		if prevaccname == accname {
			row[2] = ""
		}
		report.register = append(report.register, row)

		date, transpayee = "", ""
		prevaccname = accname
	}
	return nil
}

func (report *ReportRegister) Posting(
	db api.Datastorer, trans api.Transactor, p api.Poster) error {

	return nil
}

func (report *ReportRegister) BubblePosting(
	db api.Datastorer, trans api.Transactor,
	p api.Poster, account api.Accounter) error {

	return nil
}

func (report *ReportRegister) Render(args []string, db api.Datastorer) {
	rcf := report.rcf

	cols := []string{"By-date", "Payee", "Account", "Amount", "Balance"}
	rcf.addrow(cols...)
	rcf.addrow([]string{"", "", "", "", ""}...)

	for _, cols := range report.register {
		report.rcf.addrow(cols...)
	}

	w0 := rcf.maxwidth(rcf.column(0)) // Date
	w1 := rcf.maxwidth(rcf.column(1)) // Payee
	w2 := rcf.maxwidth(rcf.column(2)) // Account name
	w3 := rcf.maxwidth(rcf.column(3)) // Amount
	w4 := rcf.maxwidth(rcf.column(4)) // Balance (amount)
	if (w0 + w1 + w2 + w3 + w4) > 70 {
		w1 = rcf.FitPayee(1, 70-w0-w2-w3-w4)
		if (w0 + w1 + w2 + w3 + w4) > 70 {
			_ /*w2*/ = rcf.FitAccountname(1, 70-w0-w1-w3-w4)
		}
	}

	rcf.paddcells()
	fmsg := rcf.Fmsg(" %%-%vs%%-%vs%%-%vs%%%vs%%%vs\n")

	// start printing
	fmt.Println()
	for _, cols := range report.rcf.rows {
		fmt.Printf(fmsg, cols[0], cols[1], cols[2], cols[3], cols[4])
	}
	fmt.Println()
}

func (report *ReportRegister) Clone() api.Reporter {
	nreport := *report
	nreport.rcf = report.rcf.Clone()
	nreport.filteraccounts = report.filteraccounts
	nreport.filterpayees = report.filterpayees
	nreport.register = make([][]string, 0)
	return &nreport
}

func (report *ReportRegister) isfiltered() bool {
	return (len(report.filteraccounts) + len(report.filterpayees)) > 0
}
