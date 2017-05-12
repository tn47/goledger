package reports

import "strings"
import "sort"
import "fmt"

import "github.com/tn47/goledger/api"
import "github.com/tn47/goledger/dblentry"

type ReportBalance struct {
	rcf            *RCformat
	filteraccounts []string
	balance        map[string][][]string
	finaltally     [][]string
	postings       map[string]bool
	bubbleacc      map[string]bool
}

func NewReportBalance(args []string) *ReportBalance {
	report := &ReportBalance{
		rcf:       NewRCformat(),
		balance:   make(map[string][][]string),
		postings:  map[string]bool{},
		bubbleacc: map[string]bool{},
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

	report.postings[acc.Name()] = true
	return nil
}

func (report *ReportBalance) BubblePosting(
	db api.Datastorer, trans api.Transactor,
	p api.Poster, account api.Accounter) error {

	bbname := account.Name()
	// final balance
	report.finaltally = db.FmtBalances(db, trans, p, account)
	// filter account
	if api.Filterstring(bbname, report.filteraccounts) == false {
		return nil
	}
	// format account balance
	report.balance[bbname] = account.FmtBalances(db, trans, p, account)

	report.bubbleacc[bbname] = true
	return nil
}

func (report *ReportBalance) Render(db api.Datastorer, args []string) {
	report.prunebubbled()

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

func (report *ReportBalance) prunebubbled() {
	for bbname := range report.bubbleacc {
		ln, selfpost, children := len(bbname), 0, map[string]bool{}
		for postname := range report.postings {
			if postname == bbname {
				selfpost += 1
			} else if strings.HasPrefix(postname, bbname) {
				parts := dblentry.SplitAccount(postname[ln+1:])
				if postname[ln] == ':' {
					children[parts[0]] = true
				}
			}
		}
		if selfpost+len(children) < 2 {
			delete(report.balance, bbname)
		}
	}
}

func (report *ReportBalance) indent(
	indent, prefix string, keys []string) []string {

	adjustname := func(key string) {
		rows := report.balance[key]
		name := rows[len(rows)-1][1]
		if prefix != "" {
			name = strings.Trim(name[len(prefix):], ":")
			name = indent + name
		}
		rows[len(rows)-1][1] = name
	}

	newargs := func(start int, keys []string) (int, int, string, int) {
		rows := report.balance[keys[start]]
		first := rows[len(rows)-1][1] // accountname
		return start + 1, start + 1, first, len(first)
	}

	var okkeys []string

	if len(keys) > 1 {
		start, end, first, fln := newargs(0, keys)
		//fmt.Printf("enter %q %v %v\n", first, start, end)
		for end < len(keys) {
			if strings.HasPrefix(keys[end], first) && keys[end][fln] == ':' {
				end++
				continue
			}
			if end > start {
				okkeys = report.indent(indent+"  ", first, keys[start:end])
			}
			start, end, first, fln = newargs(end, keys)
		}
		if end > start {
			okkeys = report.indent(indent+"  ", first, keys[start:end])
		}
	}

	//fmt.Printf("adjustname %q %q \n", indent, prefix)
outer:
	for _, key := range keys {
		for _, okkey := range okkeys {
			if key == okkey {
				continue outer
			}
		}
		adjustname(key)
	}
	return keys
}

func (report *ReportBalance) isfiltered() bool {
	return len(report.filteraccounts) > 0
}
