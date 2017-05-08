package reports

import "strings"
import "sort"
import "fmt"

import "github.com/prataprc/goledger/api"
import "github.com/prataprc/goledger/dblentry"

type ReportBalance struct {
	rcf            *RCformat
	filteraccounts []string
	balance        map[string][][]string
	finaltally     [][]string
}

func NewReportBalance(args []string) *ReportBalance {
	report := &ReportBalance{
		rcf:     NewRCformat(),
		balance: make(map[string][][]string),
	}
	if len(args) > 1 {
		report.filteraccounts = args[1:]
	}
	return report
}

func (report *ReportBalance) Transaction(
	db api.Datastorer, trans api.Transactor) error {

	return nil
}

func (report *ReportBalance) Posting(
	db api.Datastorer, trans api.Transactor, p api.Poster) error {

	acc := p.Account()

	// final balance
	report.finaltally = db.FmtBalances(db, trans, p, acc)
	// filter account
	if api.Filterstring(acc.Name(), report.filteraccounts) == false {
		return nil
	}
	// format account balance
	report.balance[acc.Name()] = acc.FmtBalances(db, trans, p, acc)

	return nil
}

func (report *ReportBalance) BubblePosting(
	db api.Datastorer, trans api.Transactor,
	p api.Poster, account api.Accounter) error {

	children, bbname := map[string]bool{}, account.Name()
	for accname := range report.balance {
		if accname == bbname || strings.HasPrefix(accname, bbname) == false {
			continue
		}
		parts := dblentry.SplitAccount(strings.Trim(accname[len(bbname):], ":"))
		children[parts[0]] = true
	}

	if len(children) > 1 {
		// final balance
		report.finaltally = db.FmtBalances(db, trans, p, account)
		// filter account
		if api.Filterstring(bbname, report.filteraccounts) == false {
			return nil
		}
		// format account balance
		report.balance[bbname] = account.FmtBalances(db, trans, p, account)

	} else {
		delete(report.balance, bbname)
	}
	return nil
}

func (report *ReportBalance) Render(db api.Datastorer, args []string) {
	rcf := report.rcf

	// sort
	keys := []string{}
	for name := range report.balance {
		keys = append(keys, name)
	}
	sort.Strings(keys)

	report.indent("", "", keys)

	rcf.Addrow([]string{"By-date", "Account", "Balance"}...)
	rcf.Addrow([]string{"", "", ""}...) // empty line

	for _, key := range keys {
		rows := report.balance[key]
		for _, row := range rows {
			rcf.Addrow(row...)
		}
	}
	if report.isfiltered() == false {
		dashes := api.Repeatstr("-", rcf.maxwidth(rcf.column(2)))
		rcf.Addrow([]string{"", "", dashes}...)
		for _, row := range report.finaltally {
			rcf.Addrow(row...)
		}
	}

	w0 := rcf.maxwidth(rcf.column(0)) // Date
	w1 := rcf.maxwidth(rcf.column(1)) // Account name
	w2 := rcf.maxwidth(rcf.column(2)) // Balance (amount)
	if (w0 + w1 + w2) > 70 {
		w1 = rcf.FitAccountname(1, 70-w0-w2)
	}

	rcf.Paddcells()
	fmsg := rcf.Fmsg(" %%-%vs%%-%vs%%%vs\n")

	// start printing
	fmt.Println()
	for _, cols := range rcf.rows {
		fmt.Printf(fmsg, cols[0], cols[1], cols[2])
	}
	fmt.Println()
}

func (report *ReportBalance) indent(indent, prefix string, keys []string) {
	adjustname := func(key string) {
		rows := report.balance[key]
		name := rows[len(rows)-1][1]
		if prefix != "" {
			name = strings.Trim(name[len(prefix):], ":")
			name = indent + name
		}
		rows[len(rows)-1][1] = name
	}

	if len(keys) == 0 || len(keys) == 1 {
		return
	}

	newargs := func(start int) (int, int, string, int) {
		rows := report.balance[keys[start]]
		first := rows[len(rows)-1][1] // accountname
		return start + 1, start + 1, first, len(first)
	}

	start, end, first, fln := newargs(0)
	//fmt.Printf("enter %q %v %v\n", first, start, end)
	for end < len(keys) {
		if strings.HasPrefix(keys[end], first) && keys[start][fln] == ':' {
			end++
			continue
		}
		if end > start {
			report.indent(indent+"  ", first, keys[start:end])
		}
		start, end, first, fln = newargs(end)
	}
	if end > start {
		report.indent(indent+"  ", first, keys[start:end])
	}

	//fmt.Printf("adjustname %q %q \n", indent, prefix)
	for _, key := range keys {
		adjustname(key)
	}
}

func (report *ReportBalance) isfiltered() bool {
	return len(report.filteraccounts) > 0
}
