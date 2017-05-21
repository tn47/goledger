package dblentry

import "time"
import "fmt"
import "regexp"

type firstpass struct {
	defaultcomm string
	comments    []string

	currdate     time.Time
	rootaccount  string
	blncingaccnt string
	aliases      map[string]string // alias, account-alias
	payees       map[string]string // account-payee map[regex]->accountname
	repayees     map[string]*regexp.Regexp
	captures     map[string]string
	recaptures   map[string]*regexp.Regexp

	// options
	strict bool
}

func (fp *firstpass) initfirstpass() {
	fp.comments = []string{}
	fp.currdate = time.Now()
	fp.aliases = map[string]string{}
	fp.payees = map[string]string{}
	fp.repayees = map[string]*regexp.Regexp{}
	fp.captures = map[string]string{}
	fp.recaptures = map[string]*regexp.Regexp{}
}

//---- exported accessors

func (fp *firstpass) SetStrict() {
	fp.strict = true
}

func (fp *firstpass) IsStrict() bool {
	return fp.strict
}

//---- local accessors

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
	if name != "" {
		fp.blncingaccnt = name
	}
}

func (fp *firstpass) getBalancingaccount() string {
	return fp.blncingaccnt
}

func (fp *firstpass) addAlias(aliasname, accountname string) {
	if aliasname != "" {
		fp.aliases[aliasname] = accountname
	}
}

func (fp *firstpass) lookupAlias(name string) string {
	if accountname, ok := fp.aliases[name]; ok {
		return accountname
	}
	return name
}

func (fp *firstpass) addPayee(regex, accountname string) error {
	if regex != "" && accountname != "" {
		fp.payees[regex] = accountname
		regexc, err := regexp.Compile(regex)
		if err != nil {
			return err
		}
		fp.repayees[regex] = regexc
	}
	return nil
}

func (fp *firstpass) matchpayee(payee string) (string, bool) {
	for regex, regexc := range fp.repayees {
		if regexc.MatchString(payee) {
			return fp.payees[regex], true
		}
	}
	return "", false
}

func (fp *firstpass) addCapture(regex, accountname string) error {
	if regex != "" && accountname != "" {
		fp.captures[regex] = accountname
		regexc, err := regexp.Compile(regex)
		if err != nil {
			return err
		}
		fp.recaptures[regex] = regexc
	}
	return nil
}

func (fp *firstpass) matchcapture(accname string) (string, bool) {
	for regex, regexc := range fp.recaptures {
		if regexc.MatchString(accname) {
			return fp.captures[regex], true
		}
	}
	return "", false
}
