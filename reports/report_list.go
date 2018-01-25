package reports

import "fmt"

import "github.com/prataprc/goparsec"
import "github.com/bnclabs/golog"
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

func (report *ReportList) Firstpass(
	db api.Datastorer, trans api.Transactor, p api.Poster) error {

	return nil
}

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
	case "accounts", "acc":
		if api.Options.Verbose == false {
			report.listAccounts(args[2:], ndb)
		} else {
			report.listAccountsV(args[2:], ndb)
		}

	case "commodities", "commodity", "comm":
		if api.Options.Verbose == false {
			report.listCommodities(args[2:], ndb)
		} else {
			report.listCommoditiesV(args[2:], ndb)
		}
	}
}

func (report *ReportList) Clone() api.Reporter {
	nreport := *report
	nreport.rcf = report.rcf.Clone()
	return &nreport
}

func (report *ReportList) Startjournal(fname string, included bool) {
	panic("not implemented")
}

func (report *ReportList) listAccounts(args []string, ndb api.Datastorer) {
	if len(ndb.Accountnames()) == 0 {
		return
	}

	var fe *api.Filterexpr
	if len(args) > 0 {
		filterarg := api.MakeFilterexpr(args)
		node, _ := api.YFilterExpr(parsec.NewScanner([]byte(filterarg)))
		if err, ok := node.(error); ok {
			log.Errorf("filter %q expression failed: %v", filterarg, err)
			return
		}
		fe, _ = node.(*api.Filterexpr)
		log.Consolef("filter expr: %v\n", fe)
	}

	rcf := report.rcf
	for _, accname := range ndb.Accountnames() {
		if fe != nil && fe.Match(accname) == false {
			continue
		}
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
	outfd := api.Options.Outfd
	fmt.Fprintln(outfd)
	for _, cols := range rcf.rows {
		fmt.Fprintf(outfd, fmsg, cols[0], cols[1])
	}
	fmt.Fprintln(outfd)
}

func (report *ReportList) listAccountsV(args []string, ndb api.Datastorer) {
	if len(ndb.Accountnames()) == 0 {
		return
	}

	var fe *api.Filterexpr
	if len(args) > 0 {
		filterarg := api.MakeFilterexpr(args)
		node, _ := api.YFilterExpr(parsec.NewScanner([]byte(filterarg)))
		if err, ok := node.(error); ok {
			log.Errorf("filter %q expression failed: %v", filterarg, err)
			return
		}
		fe, _ = node.(*api.Filterexpr)
		//log.Consolef("filter expr: %v\n", fe)
	}

	outfd := api.Options.Outfd
	fmt.Fprintln(outfd)
	for _, accname := range ndb.Accountnames() {
		if fe != nil && fe.Match(accname) == false {
			continue
		}
		account := ndb.GetAccount(accname)
		fmt.Fprintln(outfd, account.Directive())
		fmt.Fprintln(outfd)
	}
	fmt.Fprintln(outfd)
}

func (report *ReportList) listCommodities(args []string, ndb api.Datastorer) {
	if len(ndb.Commoditynames()) == 0 {
		return
	}

	var fe *api.Filterexpr
	if len(args) > 0 {
		filterarg := api.MakeFilterexpr(args)
		node, _ := api.YFilterExpr(parsec.NewScanner([]byte(filterarg)))
		if err, ok := node.(error); ok {
			log.Errorf("filter %q expression failed: %v", filterarg, err)
			return
		}
		fe, _ = node.(*api.Filterexpr)
		//log.Consolef("filter expr: %v\n", fe)
	}

	rcf := report.rcf
	for _, commdname := range ndb.Commoditynames() {
		if fe != nil && fe.Match(commdname) == false {
			continue
		}
		commodity := ndb.GetCommodity(commdname)
		notes := commodity.Notes()
		switch len(notes) {
		case 0:
			rcf.addrow([]string{commdname, ""}...)
		case 1:
			rcf.addrow([]string{commdname, notes[0]}...)
		default:
			rcf.addrow([]string{commdname, notes[0]}...)
			for _, note := range notes[1:] {
				rcf.addrow([]string{"", note}...)
			}
		}
	}

	rcf.paddcells()
	fmsg := rcf.Fmsg(" %%-%vs%%-%vs\n")

	// start printing
	outfd := api.Options.Outfd
	fmt.Fprintln(outfd)
	for _, cols := range rcf.rows {
		fmt.Fprintf(outfd, fmsg, cols[0], cols[1])
	}
	fmt.Fprintln(outfd)
}

func (report *ReportList) listCommoditiesV(args []string, ndb api.Datastorer) {
	if len(ndb.Commoditynames()) == 0 {
		return
	}

	var fe *api.Filterexpr
	if len(args) > 0 {
		filterarg := api.MakeFilterexpr(args)
		node, _ := api.YFilterExpr(parsec.NewScanner([]byte(filterarg)))
		if err, ok := node.(error); ok {
			log.Errorf("filter %q expression failed: %v", filterarg, err)
			return
		}
		fe, _ = node.(*api.Filterexpr)
		//log.Consolef("filter expr: %v\n", fe)
	}

	outfd := api.Options.Outfd
	fmt.Fprintln(outfd)
	for _, commdname := range ndb.Commoditynames() {
		if fe != nil && fe.Match(commdname) == false {
			continue
		}
		commodity := ndb.GetCommodity(commdname)
		fmt.Fprintln(outfd, commodity.Directive())
		fmt.Fprintln(outfd)
	}
	fmt.Fprintln(outfd)
}
