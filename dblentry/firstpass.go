package dblentry

import "time"
import "fmt"

type firstpass struct {
	defaultcomm string
	comments    []string

	currdate     time.Time
	rootaccount  string
	blncingaccnt string
	aliases      map[string]string // alias, account-alias
	payees       map[string]string // account-payee map[regex]->accountname
}

func (fp *firstpass) initfirstpass() {
	fp.comments = []string{}
	fp.currdate = time.Now()
	fp.aliases = map[string]string{}
	fp.payees = map[string]string{}
}

func (fp *firstpass) setDefaultcomm(name string) {
	fp.defaultcomm = name
}

func (fp *firstpass) getDefaultcomm() string {
	return fp.defaultcomm
}

func (fp *firstpass) addComment(comment string) {
	fp.comments = append(fp.comments, comment)
}

func (fp *firstpass) setYear(year int) {
	fp.currdate = time.Date(year, 1, 1, 0, 0, 0, 0, time.Local)
}

func (fp *firstpass) getYear() int {
	return fp.currdate.Year()
}

func (fp *firstpass) setCurrentDate(date time.Time) {
	fp.currdate = date
}

func (fp *firstpass) currentDate() time.Time {
	return fp.currdate
}

func (fp *firstpass) setrootaccount(name string) error {
	if fp.rootaccount != "" {
		fmsg := "previous `apply` directive(%v) not closed"
		return fmt.Errorf(fmsg, fp.rootaccount)
	}
	fp.rootaccount = name
	return nil
}

func (fp *firstpass) clearRootaccount() error {
	if fp.rootaccount == "" {
		return fmt.Errorf("dangling `end` directive")
	}
	fp.rootaccount = ""
	return nil
}

func (fp *firstpass) applyroot(name string) string {
	if fp.rootaccount != "" {
		return fp.rootaccount + ":" + name
	}
	return name
}

func (fp *firstpass) setBalancingaccount(name string) {
	fp.blncingaccnt = name
}

func (fp *firstpass) getBalancingaccount() string {
	return fp.blncingaccnt
}

func (fp *firstpass) addAlias(aliasname, accountname string) {
	fp.aliases[aliasname] = accountname
}

func (fp *firstpass) getAlias(aliasname string) (accountname string, ok bool) {
	accountname, ok = fp.aliases[aliasname]
	return accountname, ok
}

func (fp *firstpass) lookupAlias(name string) string {
	if accountname, ok := fp.aliases[name]; ok {
		return accountname
	}
	return name
}

func (fp *Datastore) addPayee(regex, accountname string) {
	fp.payees[regex] = accountname
}
