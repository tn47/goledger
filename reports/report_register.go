package reports

import "fmt"
import "sort"

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
	balances map[string]api.Commoditiser
	lastcomm api.Commoditiser
}

// NewReportRegister create an instance for register reporting.
func NewReportRegister(args []string) (*ReportRegister, error) {
	report := &ReportRegister{
		rcf:      NewRCformat(),
		register: make([][]string, 0),
		balances: make(map[string]api.Commoditiser),
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
	for _, p := range trans.GetPostings() {
		accname := p.Account().Name()
		if report.isfilteracc() && report.fe.Match(accname) == false {
			continue
		}
		if report.isfilterpayee() && report.pfe.Match(p.Payee()) == false {
			continue
		}
		commodity := p.Commodity()
		report.applyBalance(commodity)
		row := []string{date, transpayee, accname, commodity.String(), ""}
		if p.Payee() != trans.Payee() {
			row[1] = p.Payee()
		}
		rows := report.fillbalances(row)
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

func (report *ReportRegister) isfilteracc() bool {
	return report.fe != nil
}

func (report *ReportRegister) isfilterpayee() bool {
	return report.pfe != nil
}

func (report *ReportRegister) applyBalance(comm api.Commoditiser) {
	name := comm.Name()
	if _, ok := report.balances[name]; ok == false {
		report.balances[name] = comm.MakeSimilar(0)
	}
	report.balances[name].ApplyAmount(comm)
}

func (report *ReportRegister) fillbalances(row []string) [][]string {
	if len(report.balances) == 0 {
		return [][]string{row}
	}
	names := []string{}
	for name := range report.balances {
		names = append(names, name)
	}
	sort.Strings(names)

	bal := report.balances[names[0]]
	row[len(row)-1] = bal.String()

	date, payee, accname, amount := row[0], row[1], row[2], row[3]
	rows := [][]string{}
	for _, name := range names {
		balance := report.balances[name]
		if balance.Amount() == 0 {
			continue
		}
		rw := []string{date, payee, accname, amount, balance.String()}
		rows = append(rows, rw)
		date, payee, accname, amount = "", "", "", ""
		report.lastcomm = balance
	}
	if len(rows) == 0 {
		rw := []string{date, payee, accname, amount, report.lastcomm.String()}
		rows = append(rows, rw)
	}
	return rows
}
