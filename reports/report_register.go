package reports

import "fmt"

import "github.com/prataprc/goledger/api"

type ReportRegister struct {
	rcf            *RCformat
	filteraccounts []string
	filterpayees   []string
	register       [][]string
}

func NewReportRegister(args []string) *ReportRegister {
	report := &ReportRegister{
		rcf:            NewRCformat(),
		filteraccounts: make([]string, 0),
		filterpayees:   make([]string, 0),
		register:       make([][]string, 0),
	}
	for i, arg := range args[1:] {
		if arg == "@" || arg == "payee" {
			report.filterpayees = append(report.filterpayees, args[i+1:]...)
			break
		}
		report.filteraccounts = append(report.filteraccounts, arg)
	}
	return report
}

func (report *ReportRegister) Transaction(
	db api.Datastorer, trans api.Transactor) {

	date, desc := trans.Date().Format("2006-01-02"), trans.Description()
	for _, p := range trans.GetPostings() {
		accname, payee := p.Account().Name(), desc
		if report.dofilter() {
			if api.Filterstring(accname, report.filteraccounts) == false {
				continue
			}
			if api.Filterstring(payee, report.filterpayees) == false {
				continue
			}
		}
		row := []string{
			date,
			desc,
			fmt.Sprintf("%s", accname),
			fmt.Sprintf("%.2f", p.Commodity().Amount()),
			fmt.Sprintf("%.2f", p.Account().Balance()),
		}
		report.register = append(report.register, row)
		date, desc = "", ""
	}
	return
}

func (report *ReportRegister) Posting(
	db api.Datastorer, trans api.Transactor,
	p api.Poster, account api.Accounter) {

	return
}

func (report *ReportRegister) BubblePosting(
	db api.Datastorer, trans api.Transactor,
	p api.Poster, account api.Accounter) {

	return
}

func (report *ReportRegister) Render(args []string) {
	rcf := report.rcf

	for _, cols := range report.register {
		report.rcf.Addrow(cols...)
	}

	w0 := rcf.maxwidth(rcf.column(0)) // Date
	w1 := rcf.maxwidth(rcf.column(1)) // Description
	w2 := rcf.maxwidth(rcf.column(2)) // Account name
	w3 := rcf.maxwidth(rcf.column(3)) // Amount
	w4 := rcf.maxwidth(rcf.column(4)) // Balance (amount)
	if (w0 + w1 + w2 + w3 + w4) > 70 {
		w1 = rcf.FitDescription(1, 70-w0-w2-w3-w4)
		if (w0 + w1 + w2 + w3 + w4) > 70 {
			w2 = rcf.FitAccountname(1, 70-w0-w1-w3-w4)
		}
	}

	rcf.Paddcells()
	fmsg := rcf.Fmsg(" %%-%vs%%-%vs%%-%vs%%%vs%%%vs\n")

	// start printing
	fmt.Println()
	cols := []string{
		" By-date ", " Description ", " Account ", " Amount ", " Balance ",
	}
	fmt.Printf(fmsg, cols[0], cols[1], cols[2], cols[3], cols[4])
	fmt.Println()
	for _, cols := range report.rcf.rows {
		fmt.Printf(fmsg, cols[0], cols[1], cols[2], cols[3], cols[4])
	}
	fmt.Println()
}

func (report *ReportRegister) dofilter() bool {
	return (len(report.filteraccounts) + len(report.filterpayees)) > 0
}
