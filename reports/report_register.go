package reports

import "fmt"
import "time"
import "sort"

import "github.com/prataprc/goparsec"

//import "github.com/prataprc/golog"
import "github.com/tn47/goledger/api"
import "github.com/tn47/goledger/dblentry"

// ReportRegister for register reporting.
type ReportRegister struct {
	rcf *RCformat
	fe  *api.Filterexpr
	pfe *api.Filterexpr
	// common for all map-reduce
	lastcomm api.Commoditiser
	register [][]string
	de       *dblentry.DoubleEntry
	// mapreduce-2
	begindt, enddt *time.Time
	accounts       map[string]*dblentry.DoubleEntry
	// mapreduce-3
	findates map[string]*time.Time                       // payee -> date string
	payees   map[string]map[string]*dblentry.DoubleEntry // payee -> accounts
	// mapreduce-4
	dailytm []string                                    // array of datestr
	daily   map[string]map[string]*dblentry.DoubleEntry // datestr -> accounts
	// mapreduce-5 year->week->account->de
	weeklytm map[int]map[int][2]*time.Time
	weekly   map[int]map[int]map[string]*dblentry.DoubleEntry
	// mapreduce-6 year->month->account->de
	monthly map[int]map[int]map[string]*dblentry.DoubleEntry
	// mapreduce-7 year->quarter->account->de
	quarterly map[int]map[int]map[string]*dblentry.DoubleEntry
	// mapreduce-8 year->account->de
	yearly map[int]map[string]*dblentry.DoubleEntry
}

// NewReportRegister create an instance for register reporting.
func NewReportRegister(args []string) (*ReportRegister, error) {
	report := &ReportRegister{
		rcf:       NewRCformat(),
		register:  make([][]string, 0),
		de:        dblentry.NewDoubleEntry("regbalance"),
		lastcomm:  dblentry.NewCommodity(""),
		accounts:  make(map[string]*dblentry.DoubleEntry),
		findates:  make(map[string]*time.Time),
		payees:    make(map[string]map[string]*dblentry.DoubleEntry),
		daily:     make(map[string]map[string]*dblentry.DoubleEntry),
		weeklytm:  make(map[int]map[int][2]*time.Time),
		weekly:    make(map[int]map[int]map[string]*dblentry.DoubleEntry),
		monthly:   make(map[int]map[int]map[string]*dblentry.DoubleEntry),
		quarterly: make(map[int]map[int]map[string]*dblentry.DoubleEntry),
		yearly:    make(map[int]map[string]*dblentry.DoubleEntry),
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

	if api.Options.Yearly {
		return report.mapreduce8(db, trans)
	} else if api.Options.Quarterly {
		return report.mapreduce7(db, trans)
	} else if api.Options.Monthly {
		return report.mapreduce6(db, trans)
	} else if api.Options.Weekly {
		return report.mapreduce5(db, trans)
	} else if api.Options.Daily {
		return report.mapreduce4(db, trans)
	} else if api.Options.Bypayee {
		return report.mapreduce3(db, trans)
	} else if api.Options.Subtotal {
		return report.mapreduce2(db, trans)
	}
	return report.mapreduce1(db, trans)
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
	nopayee := false
	if api.Options.Yearly {
		report.prerender9(args, db)
		nopayee = true
	} else if api.Options.Quarterly {
		report.prerender8(args, db)
		nopayee = true
	} else if api.Options.Monthly {
		report.prerender7(args, db)
		nopayee = true
	} else if api.Options.Weekly {
		report.prerender6(args, db)
		nopayee = true
	} else if api.Options.Daily {
		report.prerender5(args, db)
		nopayee = true
	} else if api.Options.Bypayee {
		report.prerender4(args, db)
	} else if api.Options.Subtotal {
		report.prerender3(args, db)
		nopayee = true
	}

	if nopayee && api.Options.Dcformat {
		report.render4(args, db)
		return
	} else if nopayee {
		report.render3(args, db)
		return
	}

	if api.Options.Dcformat == false {
		report.render1(args, db)
		return
	}
	report.render2(args, db)
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

// [-detailed] [-dc] register
func (report *ReportRegister) mapreduce1(
	db api.Datastorer, trans api.Transactor) error {

	date, transpayee := trans.Date().Format("2006-Jan-02"), trans.Payee()
	filterfn := report.matchAccOrPayee(trans)
	for _, p := range trans.GetPostings() {
		if filterfn(p) == false {
			continue
		}
		accname, comm := p.Account().Name(), p.Commodity()
		cols := []string{date, transpayee, accname}
		if api.Options.Dcformat == false {
			cols = append(cols, comm.String(), "")
		} else if comm.IsDebit() {
			cols = append(cols, comm.String(), "", "")
		} else if comm.IsCredit() {
			amount := -comm.Amount()
			cols = append(cols, "", comm.MakeSimilar(amount).String(), "")
		}
		if p.Payee() != trans.Payee() {
			cols[1] = p.Payee()
		}
		date, transpayee = "", ""
		report.de.AddBalance(comm) // should come before fillbalances
		var rows [][]string
		if api.Options.Dcformat {
			rows = report.fillbalancesDc(cols)
		} else {
			rows = report.fillbalances(cols)
		}
		report.register = append(report.register, rows...)
	}
	return nil
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

// -subtotal register
func (report *ReportRegister) mapreduce2(
	db api.Datastorer, trans api.Transactor) error {

	date := trans.Date()
	if report.begindt == nil || date.Before(*report.begindt) {
		report.begindt = &date
	}
	if report.enddt == nil || date.After(*report.enddt) {
		report.enddt = &date
	}

	filterfn := report.matchAccOrPayee(trans)
	for _, p := range trans.GetPostings() {
		if filterfn(p) == false {
			continue
		}
		accname := p.Account().Name()
		if _, ok := report.accounts[accname]; ok == false {
			report.accounts[accname] = dblentry.NewDoubleEntry(accname)
		}
		report.accounts[accname].AddBalance(p.Commodity())
	}

	return nil
}

// -bypayee register
func (report *ReportRegister) mapreduce3(
	db api.Datastorer, trans api.Transactor) error {

	date, filterfn := trans.Date(), report.matchAccOrPayee(trans)
	for _, p := range trans.GetPostings() {
		if filterfn(p) == false {
			continue
		}
		accname, payee := p.Account().Name(), p.Payee()
		_, ok := report.findates[payee]
		if ok == false || report.findates[payee].Before(date) {
			report.findates[payee] = &date
		}
		accounts, ok := report.payees[payee]
		if ok == false {
			accounts = make(map[string]*dblentry.DoubleEntry)
			report.payees[payee] = accounts
		}
		accde, ok := accounts[accname]
		if ok == false {
			accde = dblentry.NewDoubleEntry(payee + "/" + accname)
			accounts[accname] = accde
		}
		accde.AddBalance(p.Commodity())
	}
	return nil
}

// -daily register
func (report *ReportRegister) mapreduce4(
	db api.Datastorer, trans api.Transactor) error {

	datestr := trans.Date().Format("2006-Jan-02")
	filterfn := report.matchAccOrPayee(trans)
	for _, p := range trans.GetPostings() {
		if filterfn(p) == false {
			continue
		}
		accname := p.Account().Name()
		accounts, ok := report.daily[datestr]
		if ok == false {
			accounts = make(map[string]*dblentry.DoubleEntry)
			report.daily[datestr] = accounts
			report.dailytm = append(report.dailytm, datestr)
		}
		accde, ok := accounts[accname]
		if ok == false {
			accde = dblentry.NewDoubleEntry(datestr + "/" + accname)
			accounts[accname] = accde
		}
		accde.AddBalance(p.Commodity())
	}
	return nil
}

// -weekly register
func (report *ReportRegister) mapreduce5(
	db api.Datastorer, trans api.Transactor) error {

	date := trans.Date()
	year, week := trans.Date().ISOWeek()
	filterfn := report.matchAccOrPayee(trans)
	for _, p := range trans.GetPostings() {
		if filterfn(p) == false {
			continue
		}
		weeks, ok := report.weekly[year]
		if ok == false {
			weeks = make(map[int]map[string]*dblentry.DoubleEntry)
			report.weekly[year] = weeks
			report.weeklytm[year] = make(map[int][2]*time.Time)
		}
		accounts, ok := weeks[week]
		if ok == false {
			accounts = make(map[string]*dblentry.DoubleEntry)
			weeks[week] = accounts
			report.weeklytm[year][week] = [2]*time.Time{}
		}
		accname := p.Account().Name()
		accde, ok := accounts[accname]
		if ok == false {
			accde = dblentry.NewDoubleEntry(fmt.Sprintf("%v/%v", year, week))
			accounts[accname] = accde
		}
		accde.AddBalance(p.Commodity())
		be := report.weeklytm[year][week]
		if be[0] == nil || be[0].After(date) {
			be[0] = &date
		}
		if be[1] == nil || be[1].Before(date) {
			be[1] = &date
		}
		report.weeklytm[year][week] = be
	}
	return nil
}

// -monthly register
func (report *ReportRegister) mapreduce6(
	db api.Datastorer, trans api.Transactor) error {

	date := trans.Date()
	year, month := date.Year(), int(date.Month())
	filterfn := report.matchAccOrPayee(trans)
	for _, p := range trans.GetPostings() {
		if filterfn(p) == false {
			continue
		}
		months, ok := report.monthly[year]
		if ok == false {
			months = make(map[int]map[string]*dblentry.DoubleEntry)
			report.monthly[year] = months
		}
		accounts, ok := months[month]
		if ok == false {
			accounts = make(map[string]*dblentry.DoubleEntry)
			months[month] = accounts
		}
		accname := p.Account().Name()
		accde, ok := accounts[accname]
		if ok == false {
			accde = dblentry.NewDoubleEntry(fmt.Sprintf("%v/%v", year, month))
			accounts[accname] = accde
		}
		accde.AddBalance(p.Commodity())
	}
	return nil
}

// -yearly register
func (report *ReportRegister) mapreduce8(
	db api.Datastorer, trans api.Transactor) error {

	year := trans.Date().Year()
	filterfn := report.matchAccOrPayee(trans)
	for _, p := range trans.GetPostings() {
		if filterfn(p) == false {
			continue
		}
		accounts, ok := report.yearly[year]
		if ok == false {
			accounts = make(map[string]*dblentry.DoubleEntry)
			report.yearly[year] = accounts
		}
		accname := p.Account().Name()
		accde, ok := accounts[accname]
		if ok == false {
			accde = dblentry.NewDoubleEntry(fmt.Sprintf("%v", year))
			accounts[accname] = accde
		}
		accde.AddBalance(p.Commodity())
	}
	return nil
}

// -quarterly register
func (report *ReportRegister) mapreduce7(
	db api.Datastorer, trans api.Transactor) error {

	date := trans.Date()
	year, quarter := date.Year(), int(date.Month()-1)/3
	filterfn := report.matchAccOrPayee(trans)
	for _, p := range trans.GetPostings() {
		if filterfn(p) == false {
			continue
		}
		quarters, ok := report.quarterly[year]
		if ok == false {
			quarters = make(map[int]map[string]*dblentry.DoubleEntry)
			report.quarterly[year] = quarters
		}
		accounts, ok := quarters[quarter]
		if ok == false {
			accounts = make(map[string]*dblentry.DoubleEntry)
			quarters[quarter] = accounts
		}
		accname := p.Account().Name()
		accde, ok := accounts[accname]
		if ok == false {
			accde = dblentry.NewDoubleEntry(fmt.Sprintf("%v/%v", year, quarter))
			accounts[accname] = accde
		}
		accde.AddBalance(p.Commodity())
	}
	return nil
}

func (report *ReportRegister) matchAccOrPayee(
	trans api.Transactor) func(p api.Poster) bool {

	matchtrans := false
	for _, p := range trans.GetPostings() {
		accname, payee := p.Account().Name(), p.Payee()
		accok := report.isfilteracc() == false || report.fe.Match(accname)
		payeeok := report.isfilterpayee() == false || report.pfe.Match(payee)
		matchtrans = matchtrans || (accok && payeeok)
	}
	return func(p api.Poster) bool {
		if api.Options.Detailed && matchtrans {
			return true
		}
		accname, payee := p.Account().Name(), p.Payee()
		accok := report.isfilteracc() == false || report.fe.Match(accname)
		payeeok := report.isfilterpayee() == false || report.pfe.Match(payee)
		return accok && payeeok
	}
}

func (report *ReportRegister) isfilteracc() bool {
	return report.fe != nil
}

func (report *ReportRegister) isfilterpayee() bool {
	return report.pfe != nil
}

func (report *ReportRegister) render1(args []string, db api.Datastorer) {
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
	if (w0 + w1 + w2 + w3 + w4) > 125 {
		w1 = rcf.FitPayee(1, 125-w0-w2-w3-w4)
		if (w0 + w1 + w2 + w3 + w4) > 125 {
			_ /*w2*/ = rcf.FitAccountname(1, 125-w0-w1-w3-w4)
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

func (report *ReportRegister) render2(args []string, db api.Datastorer) {
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
			_ /*w2*/ = rcf.FitAccountname(1, 125-w0-w1-w3-w4-w5)
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

// nopayee
func (report *ReportRegister) render3(args []string, db api.Datastorer) {
	rcf := report.rcf

	cols := []string{"By-date", "Account", "Amount", "Balance"}
	rcf.addrow(cols...)
	rcf.addrow([]string{"", "", "", ""}...)

	for _, cols := range report.register {
		report.rcf.addrow(cols...)
	}

	w0 := rcf.maxwidth(rcf.column(0)) // Date
	w1 := rcf.maxwidth(rcf.column(1)) // Account name
	w2 := rcf.maxwidth(rcf.column(2)) // Amount
	w3 := rcf.maxwidth(rcf.column(3)) // Balance (amount)
	if (w0 + w1 + w2 + w3) > 125 {
		w1 = rcf.FitAccountname(1, 125-w0-w2-w3)
	}

	rcf.paddcells()
	fmsg := rcf.Fmsg(" %%-%vs%%-%vs%%%vs%%%vs\n")
	comm1 := dblentry.NewCommodity("")
	comm2 := dblentry.NewCommodity("")

	// start printing
	outfd := api.Options.Outfd
	fmt.Fprintln(outfd)
	for i, cols := range report.rcf.rows {
		items := []interface{}{cols[0]}
		if i < 2 {
			items = append(items, cols[1], cols[2], cols[3])
		} else {
			x := CommodityColor(db, comm1, cols[2])
			y := CommodityColor(db, comm2, cols[3])
			items = append(items, api.YellowFn(cols[1]), x, y)
		}
		fmt.Fprintf(outfd, fmsg, items...)
	}
	fmt.Fprintln(outfd)
}

// nopayee, dcformat
func (report *ReportRegister) render4(args []string, db api.Datastorer) {
	rcf := report.rcf

	cols := []string{"By-date", "Account", "Debit", "Credit", "Balance"}
	rcf.addrow(cols...)
	rcf.addrow([]string{"", "", "", "", ""}...)

	for _, cols := range report.register {
		report.rcf.addrow(cols...)
	}

	w0 := rcf.maxwidth(rcf.column(0)) // Date
	w1 := rcf.maxwidth(rcf.column(1)) // Account name
	w2 := rcf.maxwidth(rcf.column(2)) // Debit
	w3 := rcf.maxwidth(rcf.column(3)) // Credit
	w4 := rcf.maxwidth(rcf.column(4)) // Balance (amount)
	if (w0 + w1 + w2 + w3 + w4) > 125 {
		w1 = rcf.FitAccountname(1, 125-w0-w2-w3-w4)
	}

	rcf.paddcells()
	fmsg := rcf.Fmsg(" %%-%vs%%-%vs%%-%vs%%%vs%%%vs\n")
	comm1 := dblentry.NewCommodity("")

	// start printing
	outfd := api.Options.Outfd
	fmt.Fprintln(outfd)
	for i, cols := range report.rcf.rows {
		items := []interface{}{cols[0]}
		if i < 2 {
			items = append(items, cols[1], cols[2], cols[3], cols[4])
		} else {
			x := CommodityColor(db, comm1, cols[4])
			items = append(items, api.YellowFn(cols[1]), cols[2], cols[3], x)
		}
		fmt.Fprintf(outfd, fmsg, items...)
	}
	fmt.Fprintln(outfd)
}

// -subtotal
func (report *ReportRegister) prerender3(args []string, db api.Datastorer) {
	accnames := []string{}
	for accname := range report.accounts {
		accnames = append(accnames, accname)
	}
	sort.Strings(accnames)

	report.de = dblentry.NewDoubleEntry("regbalance")
	report.register = [][]string{}
	for _, accname := range accnames {
		de, rows, balnames := report.accounts[accname], [][]string{}, []string{}
		for _, abal := range de.Balances() {
			cols := []string{"", ""} // date-range, accname
			if api.Options.Dcformat == false {
				cols = append(cols, abal.String(), "")
			} else if abal.IsDebit() {
				cols = append(cols, abal.String(), "", "")
			} else if abal.IsCredit() {
				abalstr := abal.MakeSimilar(-abal.Amount()).String()
				cols = append(cols, "", abalstr, "")
			}
			report.de.AddBalance(abal)
			cols[len(cols)-1] = report.de.Balance(abal.Name()).String()
			rows, balnames = append(rows, cols), append(balnames, abal.Name())
		}
		for _, bal := range report.de.Balances() {
			if api.HasString(balnames, bal.Name()) {
				continue
			}
			if api.Options.Dcformat {
				rows = append(rows, []string{"", "", "", "", bal.String()})
			} else {
				rows = append(rows, []string{"", "", "", bal.String()})
			}
		}
		rows[0][1] = accname
		report.register = append(report.register, rows...)
	}
	if len(report.register) > 0 {
		x := report.begindt.Format("2006-Jan-02")
		y := report.enddt.Format("2006-Jan-02")
		report.register[0][0] = fmt.Sprintf("%v to %v", x, y)
	}
}

// -bypayee
func (report *ReportRegister) prerender4(args []string, db api.Datastorer) {
	payees := []string{}
	for payee := range report.payees {
		payees = append(payees, payee)
	}
	sort.Strings(payees)

	sortaccount := func(accounts map[string]*dblentry.DoubleEntry) []string {
		accnames := []string{}
		for accname := range accounts {
			accnames = append(accnames, accname)
		}
		sort.Strings(accnames)
		return accnames
	}

	report.de = dblentry.NewDoubleEntry("regbalance")
	report.register = [][]string{}
	for _, payee := range payees {
		payeerows, balrows, balnames := [][]string{}, [][]string{}, []string{}
		accnames := sortaccount(report.payees[payee])
		for _, accname := range accnames {
			de := report.payees[payee][accname]
			accrows := [][]string{}
			for _, abal := range de.Balances() {
				cols := []string{"", "", ""} // date-range, payee, accname
				if api.Options.Dcformat == false {
					cols = append(cols, abal.String(), "")
				} else if abal.IsDebit() {
					cols = append(cols, abal.String(), "", "")
				} else if abal.IsCredit() {
					abalstr := abal.MakeSimilar(-abal.Amount()).String()
					cols = append(cols, "", abalstr, "")
				}
				report.de.AddBalance(abal)
				cols[len(cols)-1] = report.de.Balance(abal.Name()).String()
				accrows = append(accrows, cols)
				balnames = append(balnames, abal.Name())
			}
			accrows[0][2] = accname
			balrows = append(balrows, accrows...)
		}
		payeerows = append(payeerows, balrows...)

		for _, bal := range report.de.Balances() {
			if api.HasString(balnames, bal.Name()) {
				continue
			}
			var cols []string
			if api.Options.Dcformat {
				cols = []string{"", "", "", "", "", bal.String()}
			} else {
				cols = []string{"", "", "", "", bal.String()}
			}
			payeerows = append(payeerows, cols)
		}

		payeerows[0][0] = report.findates[payee].Format("2006-Jan-02")
		payeerows[0][1] = payee
		report.register = append(report.register, payeerows...)
	}
}

// -daily
func (report *ReportRegister) prerender5(args []string, db api.Datastorer) {
	sortaccount := func(accounts map[string]*dblentry.DoubleEntry) []string {
		accnames := []string{}
		for accname := range accounts {
			accnames = append(accnames, accname)
		}
		sort.Strings(accnames)
		return accnames
	}

	report.de = dblentry.NewDoubleEntry("regbalance")
	report.register = [][]string{}
	for _, datestr := range report.dailytm {
		daterows, balrows, balnames := [][]string{}, [][]string{}, []string{}
		accnames := sortaccount(report.daily[datestr])
		for _, accname := range accnames {
			de := report.daily[datestr][accname]
			accrows := [][]string{}
			for _, abal := range de.Balances() {
				cols := []string{"", ""} // date-range, accname
				if api.Options.Dcformat == false {
					cols = append(cols, abal.String(), "")
				} else if abal.IsDebit() {
					cols = append(cols, abal.String(), "", "")
				} else if abal.IsCredit() {
					abalstr := abal.MakeSimilar(-abal.Amount()).String()
					cols = append(cols, "", abalstr, "")
				}
				report.de.AddBalance(abal)
				cols[len(cols)-1] = report.de.Balance(abal.Name()).String()
				accrows = append(accrows, cols)
				balnames = append(balnames, abal.Name())
			}
			accrows[0][1] = accname
			balrows = append(balrows, accrows...)
		}
		daterows = append(daterows, balrows...)

		for _, bal := range report.de.Balances() {
			if api.HasString(balnames, bal.Name()) {
				continue
			}
			var cols []string
			if api.Options.Dcformat {
				cols = []string{"", "", "", "", bal.String()}
			} else {
				cols = []string{"", "", "", bal.String()}
			}
			daterows = append(daterows, cols)
		}

		daterows[0][0] = datestr
		report.register = append(report.register, daterows...)
	}
}

// -weekly
func (report *ReportRegister) prerender6(args []string, db api.Datastorer) {
	years := []int{}
	for year := range report.weekly {
		years = append(years, year)
	}
	sort.Ints(years)

	sortaccount := func(accounts map[string]*dblentry.DoubleEntry) []string {
		accnames := []string{}
		for accname := range accounts {
			accnames = append(accnames, accname)
		}
		sort.Strings(accnames)
		return accnames
	}
	sortweeks := func(weeks map[int]map[string]*dblentry.DoubleEntry) []int {
		weekns := []int{}
		for week := range weeks {
			weekns = append(weekns, week)
		}
		sort.Ints(weekns)
		return weekns
	}

	report.de = dblentry.NewDoubleEntry("regbalance")
	report.register = [][]string{}
	for _, year := range years {
		weeks := report.weekly[year]
		weekns := sortweeks(weeks)
		for _, week := range weekns {
			accounts := weeks[week]
			daterows, balrows, balnames := [][]string{}, [][]string{}, []string{}
			accnames := sortaccount(accounts)
			for _, accname := range accnames {
				de, accrows := accounts[accname], [][]string{}
				for _, abal := range de.Balances() {
					cols := []string{"", ""} // date-range, accname
					if api.Options.Dcformat == false {
						cols = append(cols, abal.String(), "")
					} else if abal.IsDebit() {
						cols = append(cols, abal.String(), "", "")
					} else if abal.IsCredit() {
						abalstr := abal.MakeSimilar(-abal.Amount()).String()
						cols = append(cols, "", abalstr, "")
					}
					report.de.AddBalance(abal)
					cols[len(cols)-1] = report.de.Balance(abal.Name()).String()
					accrows = append(accrows, cols)
					balnames = append(balnames, abal.Name())
				}
				accrows[0][1] = accname
				balrows = append(balrows, accrows...)
			}
			daterows = append(daterows, balrows...)

			for _, bal := range report.de.Balances() {
				if api.HasString(balnames, bal.Name()) {
					continue
				}
				var cols []string
				if api.Options.Dcformat {
					cols = []string{"", "", "", "", bal.String()}
				} else {
					cols = []string{"", "", "", bal.String()}
				}
				daterows = append(daterows, cols)
			}

			be := report.weeklytm[year][week]
			x, y := be[0].Format("2006-Jan-02"), be[1].Format("2006-Jan-02")
			daterows[0][0] = fmt.Sprintf("%v - %v", x, y)
			report.register = append(report.register, daterows...)
		}
	}
}

// -monthly register
func (report *ReportRegister) prerender7(args []string, db api.Datastorer) {
	years := []int{}
	for year := range report.monthly {
		years = append(years, year)
	}
	sort.Ints(years)

	sortaccount := func(accounts map[string]*dblentry.DoubleEntry) []string {
		accnames := []string{}
		for accname := range accounts {
			accnames = append(accnames, accname)
		}
		sort.Strings(accnames)
		return accnames
	}
	sortmonths := func(months map[int]map[string]*dblentry.DoubleEntry) []int {
		monthns := []int{}
		for month := range months {
			monthns = append(monthns, month)
		}
		sort.Ints(monthns)
		return monthns
	}

	report.de = dblentry.NewDoubleEntry("regbalance")
	report.register = [][]string{}
	for _, year := range years {
		months := report.monthly[year]
		monthns := sortmonths(months)
		for _, month := range monthns {
			accounts := months[month]
			daterows, balrows, balnames := [][]string{}, [][]string{}, []string{}
			accnames := sortaccount(accounts)
			for _, accname := range accnames {
				de, accrows := accounts[accname], [][]string{}
				for _, abal := range de.Balances() {
					cols := []string{"", ""} // date-range, accname
					if api.Options.Dcformat == false {
						cols = append(cols, abal.String(), "")
					} else if abal.IsDebit() {
						cols = append(cols, abal.String(), "", "")
					} else if abal.IsCredit() {
						abalstr := abal.MakeSimilar(-abal.Amount()).String()
						cols = append(cols, "", abalstr, "")
					}
					report.de.AddBalance(abal)
					cols[len(cols)-1] = report.de.Balance(abal.Name()).String()
					accrows = append(accrows, cols)
					balnames = append(balnames, abal.Name())
				}
				accrows[0][1] = accname
				balrows = append(balrows, accrows...)
			}
			daterows = append(daterows, balrows...)

			for _, bal := range report.de.Balances() {
				if api.HasString(balnames, bal.Name()) {
					continue
				}
				var cols []string
				if api.Options.Dcformat {
					cols = []string{"", "", "", "", bal.String()}
				} else {
					cols = []string{"", "", "", bal.String()}
				}
				daterows = append(daterows, cols)
			}

			tm := time.Date(year, time.Month(month), 0, 0, 0, 0, 0, time.Local)
			daterows[0][0] = tm.Format("2006-Jan")
			report.register = append(report.register, daterows...)
		}
	}
}

// -quarterly register
func (report *ReportRegister) prerender8(args []string, db api.Datastorer) {
	years := []int{}
	for year := range report.quarterly {
		years = append(years, year)
	}
	sort.Ints(years)

	sortaccount := func(accounts map[string]*dblentry.DoubleEntry) []string {
		accnames := []string{}
		for accname := range accounts {
			accnames = append(accnames, accname)
		}
		sort.Strings(accnames)
		return accnames
	}
	sortquarters := func(qs map[int]map[string]*dblentry.DoubleEntry) []int {
		qns := []int{}
		for q := range qs {
			qns = append(qns, q)
		}
		sort.Ints(qns)
		return qns
	}

	report.de = dblentry.NewDoubleEntry("regbalance")
	report.register = [][]string{}
	for _, year := range years {
		quarters := report.quarterly[year]
		qns := sortquarters(quarters)
		for _, quarter := range qns {
			accounts := quarters[quarter]
			daterows, balrows, balnames := [][]string{}, [][]string{}, []string{}
			accnames := sortaccount(accounts)
			for _, accname := range accnames {
				de, accrows := accounts[accname], [][]string{}
				for _, abal := range de.Balances() {
					cols := []string{"", ""} // date-range, accname
					if api.Options.Dcformat == false {
						cols = append(cols, abal.String(), "")
					} else if abal.IsDebit() {
						cols = append(cols, abal.String(), "", "")
					} else if abal.IsCredit() {
						abalstr := abal.MakeSimilar(-abal.Amount()).String()
						cols = append(cols, "", abalstr, "")
					}
					report.de.AddBalance(abal)
					cols[len(cols)-1] = report.de.Balance(abal.Name()).String()
					accrows = append(accrows, cols)
					balnames = append(balnames, abal.Name())
				}
				accrows[0][1] = accname
				balrows = append(balrows, accrows...)
			}
			daterows = append(daterows, balrows...)

			for _, bal := range report.de.Balances() {
				if api.HasString(balnames, bal.Name()) {
					continue
				}
				var cols []string
				if api.Options.Dcformat {
					cols = []string{"", "", "", "", bal.String()}
				} else {
					cols = []string{"", "", "", bal.String()}
				}
				daterows = append(daterows, cols)
			}

			daterows[0][0] = fmt.Sprintf("%v/q%v", year, quarter+1)
			report.register = append(report.register, daterows...)
		}
	}
}

// -yearly register
func (report *ReportRegister) prerender9(args []string, db api.Datastorer) {
	years := []int{}
	for year := range report.yearly {
		years = append(years, year)
	}
	sort.Ints(years)

	sortaccount := func(accounts map[string]*dblentry.DoubleEntry) []string {
		accnames := []string{}
		for accname := range accounts {
			accnames = append(accnames, accname)
		}
		sort.Strings(accnames)
		return accnames
	}

	report.de = dblentry.NewDoubleEntry("regbalance")
	report.register = [][]string{}
	for _, year := range years {
		accounts := report.yearly[year]
		daterows, balrows, balnames := [][]string{}, [][]string{}, []string{}
		accnames := sortaccount(accounts)
		for _, accname := range accnames {
			de, accrows := accounts[accname], [][]string{}
			for _, abal := range de.Balances() {
				cols := []string{"", ""} // date-range, accname
				if api.Options.Dcformat == false {
					cols = append(cols, abal.String(), "")
				} else if abal.IsDebit() {
					cols = append(cols, abal.String(), "", "")
				} else if abal.IsCredit() {
					abalstr := abal.MakeSimilar(-abal.Amount()).String()
					cols = append(cols, "", abalstr, "")
				}
				report.de.AddBalance(abal)
				cols[len(cols)-1] = report.de.Balance(abal.Name()).String()
				accrows = append(accrows, cols)
				balnames = append(balnames, abal.Name())
			}
			accrows[0][1] = accname
			balrows = append(balrows, accrows...)
		}
		daterows = append(daterows, balrows...)

		for _, bal := range report.de.Balances() {
			if api.HasString(balnames, bal.Name()) {
				continue
			}
			var cols []string
			if api.Options.Dcformat {
				cols = []string{"", "", "", "", bal.String()}
			} else {
				cols = []string{"", "", "", bal.String()}
			}
			daterows = append(daterows, cols)
		}

		daterows[0][0] = fmt.Sprintf("%v", year)
		report.register = append(report.register, daterows...)
	}
}
