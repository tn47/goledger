package reports

import "fmt"

import "github.com/prataprc/goparsec"

//import "github.com/prataprc/golog"
import "github.com/tn47/goledger/api"
import "github.com/tn47/goledger/dblentry"

// ReportRegister for register reporting.
type ReportRegister struct {
	rcf      *RCformat
	fe       *api.Filterexpr
	pfe      *api.Filterexpr
	register [][]string
	de       *dblentry.DoubleEntry
	lastcomm api.Commoditiser
}

// NewReportRegister create an instance for register reporting.
func NewReportRegister(args []string) (*ReportRegister, error) {
	report := &ReportRegister{
		rcf:      NewRCformat(),
		register: make([][]string, 0),
		de:       dblentry.NewDoubleEntry("registerbalance"),
		lastcomm: dblentry.NewCommodity(""),
	}

	filteraccounts := []string{}
	for i, arg := range args[1:] {
		if arg == "@" || arg == "payee" && len(args[i:]) > 1 {
			filterarg := api.MakeFilterexpr(args[i+1:])
			node, _ := api.YFilterExpr(parsec.NewScanner([]byte(filterarg)))
			if err, ok := node.(error); ok {
				return nil, err
			}
			report.pfe = node.(*api.Filterexpr)
			//log.Consolef("filter expr: %v\n", report.pfe)
			break
		}
		filteraccounts = append(filteraccounts, arg)
	}
	if len(filteraccounts) > 0 {
		filterarg := api.MakeFilterexpr(filteraccounts)
		node, _ := api.YFilterExpr(parsec.NewScanner([]byte(filterarg)))
		if err, ok := node.(error); ok {
			return nil, err
		}
		report.fe = node.(*api.Filterexpr)
		//log.Consolef("filter expr: %v\n", report.fe)
	}
	return report, nil
}

//---- api.Reporter methods

func (report *ReportRegister) Firstpass(
	db api.Datastorer, trans api.Transactor, p api.Poster) error {

	return nil
}

func (report *ReportRegister) Transaction(
	db api.Datastorer, trans api.Transactor) error {

	if api.FilterPeriod(trans.Date(), false /*nobegin*/) == false {
		return nil
	}

	date, transpayee := trans.Date().Format("2006-Jan-02"), trans.Payee()
	matchok := false // short-circuit for matchForDetail()
	for _, p := range trans.GetPostings() {
		var cols []string
		var ok bool

		accname, payee, comm := p.Account().Name(), p.Payee(), p.Commodity()

		ok, matchok = report.matchForDetail(matchok, accname, payee)
		if ok == false {
			continue
		}

		amountstr := comm.String()
		report.de.AddBalance(comm)
		if api.Options.Dcformat == false {
			cols = []string{date, transpayee, accname, amountstr, ""}
		} else if comm.IsDebit() {
			cols = []string{date, transpayee, accname, amountstr, "", ""}
		} else if comm.IsCredit() {
			amountstr = comm.MakeSimilar(-comm.Amount()).String()
			cols = []string{date, transpayee, accname, "", amountstr, ""}
		}
		if p.Payee() != trans.Payee() {
			cols[1] = p.Payee()
		}
		var rows [][]string
		if api.Options.Dcformat {
			rows = report.fillbalancesDc(cols)
		} else {
			rows = report.fillbalances(cols)
		}
		report.register = append(report.register, rows...)

		date, transpayee = "", ""
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
	if api.Options.Dcformat {
		report.renderDcRegister(args, db)
	} else {
		report.renderRegister(args, db)
	}
}

func (report *ReportRegister) renderRegister(
	args []string, db api.Datastorer) {

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
	comm1 := dblentry.NewCommodity("")
	comm2 := dblentry.NewCommodity("")

	// start printing
	outfd := api.Options.Outfd
	fmt.Fprintln(outfd)
	for i, cols := range report.rcf.rows {
		items := []interface{}{cols[0], cols[1]}
		if i < 2 {
			items = append(items, cols[2], cols[3], cols[4])
		} else {
			x := CommodityColor(db, comm1, cols[3])
			y := CommodityColor(db, comm2, cols[4])
			items = append(items, api.YellowFn(cols[2]), x, y)
		}
		fmt.Fprintf(outfd, fmsg, items...)
	}
	fmt.Fprintln(outfd)
}

func (report *ReportRegister) renderDcRegister(
	args []string, db api.Datastorer) {

	rcf := report.rcf

	cols := []string{"By-date", "Payee", "Account", "Debit", "Credit", "Balance"}
	rcf.addrow(cols...)
	rcf.addrow([]string{"", "", "", "", "", ""}...)

	for _, cols := range report.register {
		report.rcf.addrow(cols...)
	}

	w0 := rcf.maxwidth(rcf.column(0)) // Date
	w1 := rcf.maxwidth(rcf.column(1)) // Payee
	w2 := rcf.maxwidth(rcf.column(2)) // Account name
	w3 := rcf.maxwidth(rcf.column(3)) // Debit
	w4 := rcf.maxwidth(rcf.column(4)) // Credit
	w5 := rcf.maxwidth(rcf.column(5)) // Balance (amount)
	if (w0 + w1 + w2 + w3 + w4 + w5) > 125 {
		w1 = rcf.FitPayee(1, 125-w0-w2-w3-w4-w5)
		if (w0 + w1 + w2 + w3 + w4 + w5) > 125 {
			_ /*w2*/ = rcf.FitAccountname(1, 70-w0-w1-w3-w4-w5)
		}
	}

	rcf.paddcells()
	fmsg := rcf.Fmsg(" %%-%vs%%-%vs%%-%vs%%%vs%%%vs%%%vs\n")
	comm1 := dblentry.NewCommodity("")

	// start printing
	outfd := api.Options.Outfd
	fmt.Fprintln(outfd)
	for i, cols := range report.rcf.rows {
		items := []interface{}{cols[0], cols[1]}
		if i < 2 {
			items = append(items, cols[2], cols[3], cols[4], cols[5])
		} else {
			x := CommodityColor(db, comm1, cols[5])
			items = append(items, api.YellowFn(cols[2]), cols[3], cols[4], x)
		}
		fmt.Fprintf(outfd, fmsg, items...)
	}
	fmt.Fprintln(outfd)
}

func (report *ReportRegister) Clone() api.Reporter {
	nreport := *report
	nreport.rcf = report.rcf.Clone()
	nreport.pfe = report.pfe
	nreport.fe = report.fe
	nreport.register = make([][]string, 0)
	return &nreport
}

func (report *ReportRegister) Startjournal(fname string, included bool) {
	panic("not implemented")
}

func (report *ReportRegister) fillbalances(cols []string) [][]string {
	balances := report.de.Balances()
	if len(balances) == 0 {
		return [][]string{cols}
	}

	date, payee, accname, amount := cols[0], cols[1], cols[2], cols[3]
	rows := [][]string{}
	for _, balance := range balances {
		if balance.Amount() == 0 {
			continue
		}
		cols := []string{date, payee, accname, amount, balance.String()}
		rows = append(rows, cols)
		date, payee, accname, amount = "", "", "", ""
		report.lastcomm = balance
	}
	if len(rows) == 0 {
		cols := []string{date, payee, accname, amount, report.lastcomm.String()}
		rows = append(rows, cols)
	}
	return rows
}

func (report *ReportRegister) fillbalancesDc(cols []string) [][]string {
	balances := report.de.Balances()
	if len(balances) == 0 {
		return [][]string{cols}
	}

	date, payee, accname, dr, cr := cols[0], cols[1], cols[2], cols[3], cols[4]
	rows := [][]string{}
	for _, balance := range balances {
		if balance.Amount() == 0 {
			continue
		}
		cols := []string{date, payee, accname, dr, cr, balance.String()}
		rows = append(rows, cols)
		date, payee, accname, dr, cr = "", "", "", "", ""
		report.lastcomm = balance
	}
	if len(rows) == 0 {
		cols := []string{date, payee, accname, dr, cr, report.lastcomm.String()}
		rows = append(rows, cols)
	}
	return rows
}

func (report *ReportRegister) matchForDetail(
	matchok bool, accname, payee string) (bool, bool) {

	if api.Options.Detailed && matchok {
		return true, true
	}
	if report.isfilteracc() {
		if report.fe.Match(accname) {
			return true, true
		}
		return false, false
	}
	if report.isfilterpayee() {
		if report.pfe.Match(payee) {
			return true, true
		}
		return false, false
	}
	return true, false
}

func (report *ReportRegister) isfilteracc() bool {
	return report.fe != nil
}

func (report *ReportRegister) isfilterpayee() bool {
	return report.pfe != nil
}
