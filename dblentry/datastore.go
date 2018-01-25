package dblentry

import "fmt"
import "sort"
import "time"
import "strings"

import "github.com/prataprc/goparsec"
import "github.com/bnclabs/golog"
import "github.com/tn47/goledger/api"

// ParsePhase state to be tracked at datastore.
type ParsePhase int

const (
	// DBSTART is parsing started.
	DBSTART ParsePhase = iota + 1
	// DBFIRSTPASS is all Firstpass() calls are completed.
	DBFIRSTPASS
	// DBSECONDPASS is all Secondpass() calls are completed.
	DBSECONDPASS
)

// Datastore managing accounts, commodities, transactions and posting and
// every thing else that are related.
type Datastore struct {
	// immutable from first initialization.
	name string

	// immutable once firstpass is ok
	journals    map[uint64]string
	currjournal string
	firstpass

	// changes with every second pass.
	reporter    api.Reporter
	pass        ParsePhase
	commodities map[string]*Commodity
	accntdb     map[string]*Account // full account-name -> account
	dclrdacc    []string
	dclrdcomm   []string
	dclrdpayee  []string
	de          *DoubleEntry
	transdb     *DB
	pricedb     *DB

	// configuration
	periodtill *time.Time
}

// NewDatastore return a new datastore.
func NewDatastore(name string, reporter api.Reporter) *Datastore {
	db := &Datastore{
		name:     name,
		journals: make(map[uint64]string),
		reporter: reporter,

		pass:        DBSTART,
		transdb:     NewDB(fmt.Sprintf("%v-transactions", name)),
		pricedb:     NewDB(fmt.Sprintf("%v-pricedb", name)),
		accntdb:     map[string]*Account{},
		commodities: map[string]*Commodity{},
		de:          NewDoubleEntry("master"),
	}
	db.initfirstpass()
	db.defaultprices()
	return db
}

//---- local accessors

func (db *Datastore) assertfirstpass() {
	if db.pass < DBFIRSTPASS {
		panic("impossible situation")
	}
}

func (db *Datastore) getCommodity(name string, defcomm *Commodity) *Commodity {
	defaultcomm := db.getDefaultcomm()
	if name == "" && defaultcomm != "" {
		return db.commodities[defaultcomm]
	}
	if defaultcomm == "" && name != "" {
		db.setDefaultcomm(name)
	}
	if comm, ok := db.commodities[name]; ok {
		return comm
	}
	log.Debugf("commodity %q added\n", name)
	if defcomm == nil {
		defcomm = NewCommodity(name)
	}
	db.commodities[name] = defcomm
	return defcomm
}

//---- exported accessors

func (db *Datastore) Addjournal(journalfile string, data []byte) {
	hash := api.Crc64(data)
	db.journals[hash] = journalfile
	db.currjournal = journalfile
}

func (db *Datastore) Hasjournal(data []byte) bool {
	hash := api.Crc64(data)
	_, ok := db.journals[hash]
	return ok
}

func (db *Datastore) CurrentJournal() string {
	return db.currjournal
}

func (db *Datastore) GetCommodity(name string) api.Commoditiser {
	return db.getCommodity(name, nil)
}

func (db *Datastore) Applytill(till time.Time) {
	periodtill := till
	db.periodtill = &periodtill
}

// Firstpassok to track parsephase
func (db *Datastore) Firstpassok() {
	db.pass = DBFIRSTPASS
}

// Secondpassok to track parsephase
func (db *Datastore) Secondpassok() {
	db.pass = DBSECONDPASS
}

// PrintAccounts is a debug api for application.
func (db *Datastore) PrintAccounts() {
	for _, accname := range db.Accountnames() {
		log.Debugf("-- %v\n", db.accntdb[accname])
	}
}

//---- api.Datastorer methods

func (db *Datastore) IsCommodityDeclared(name string) bool {
	for _, xname := range db.dclrdcomm {
		if xname == name {
			return true
		}
	}
	return false
}

func (db *Datastore) IsAccountDeclared(name string) bool {
	for _, xname := range db.dclrdacc {
		if xname == name {
			return true
		}
	}
	return false
}

func (db *Datastore) IsPayeeDeclared(name string) bool {
	for _, xname := range db.dclrdpayee {
		if xname == name {
			return true
		}
	}
	return false
}

func (db *Datastore) GetAccount(name string) api.Accounter {
	if name == "" {
		return (*Account)(nil)
	}
	account, ok := db.accntdb[name]
	if ok == false {
		account = NewAccount(name)
	}
	db.accntdb[name] = account
	return account
}

func (db *Datastore) Accountnames() []string {
	db.assertfirstpass()

	accnames := []string{}
	for name := range db.accntdb {
		accnames = append(accnames, name)
	}
	sort.Strings(accnames)
	return accnames
}

func (db *Datastore) Commoditynames() []string {
	db.assertfirstpass()

	cnames := []string{}
	for name := range db.commodities {
		cnames = append(cnames, name)
	}
	sort.Strings(cnames)
	return cnames
}

func (db *Datastore) Balance(obj interface{}) (balance api.Commoditiser) {
	db.assertfirstpass()
	return db.de.Balance(obj)
}

func (db *Datastore) Balances() []api.Commoditiser {
	db.assertfirstpass()
	return db.de.Balances()
}

func (db *Datastore) Debit(obj interface{}) (balance api.Commoditiser) {
	db.assertfirstpass()
	return db.de.Debit(obj)
}

func (db *Datastore) Debits() []api.Commoditiser {
	db.assertfirstpass()
	return db.de.Debits()
}

func (db *Datastore) Credit(obj interface{}) (balance api.Commoditiser) {
	db.assertfirstpass()
	return db.de.Credit(obj)
}

func (db *Datastore) Credits() []api.Commoditiser {
	db.assertfirstpass()
	return db.de.Credits()
}

func (db *Datastore) AggregateTotal(trans api.Transactor, p api.Poster) error {
	posting := p.(*Posting)
	names := SplitAccount(posting.account.name)
	parts := []string{}
	for _, name := range names[:len(names)-1] {
		parts = append(parts, name)
		fullname := strings.Join(parts, ":")
		consacc := db.GetAccount(fullname).(*Account)
		if err := consacc.addBalance(posting.commodity); err != nil {
			return err
		}
		err := db.reporter.BubblePosting(db, trans, posting, consacc)
		if err != nil {
			return err
		}
	}
	return nil
}

//---- engine

func (db *Datastore) Firstpass(obj interface{}) (err error) {
	if trans, ok := obj.(*Transaction); ok {
		if err := trans.Firstpass(db); err != nil {
			return err
		}
		db.setCurrentDate(trans.date)
		db.transdb.Insert(trans.date, trans)

	} else if price, ok := obj.(*Price); ok {
		err = db.pricedb.Insert(price.when, price)

	} else if directive, ok := obj.(*Directive); ok {
		err = directive.Firstpass(db)

	} else if comment, ok := obj.(*Comment); ok {
		err = comment.Firstpass(db)
		db.addComment(comment.line)
	}
	return err
}

func (db *Datastore) Secondpass() error {
	entries := []api.TimeEntry{}

	for _, entry := range db.transdb.Range(nil, nil, "both", entries) {
		trans := entry.Value().(*Transaction)
		if db.periodtill == nil || trans.Date().Before(*db.periodtill) {
			if err := trans.Secondpass(db); err != nil {
				return err
			}
		}
	}
	return nil
}

func (db *Datastore) Clone(nreporter api.Reporter) api.Datastorer {
	ndb := *db
	ndb.reporter = nreporter

	ndb.commodities = map[string]*Commodity{}
	for name, commodity := range db.commodities {
		ndb.commodities[name] = commodity.Clone(&ndb).(*Commodity)
	}

	ndb.accntdb = map[string]*Account{}
	for name, account := range db.accntdb {
		ndb.accntdb[name] = account.Clone(&ndb)
	}

	ndb.de = db.de.Clone()

	ndb.transdb = NewDB(fmt.Sprintf("%v-transactions", ndb.name))
	entries := []api.TimeEntry{}
	for _, entry := range db.transdb.Range(nil, nil, "both", entries) {
		k, ntrans := entry.Key(), entry.Value().(*Transaction).Clone(&ndb)
		ndb.transdb.Insert(k, ntrans)
	}

	ndb.pricedb = NewDB(fmt.Sprintf("%v-pricedb", ndb.name))
	entries = []api.TimeEntry{}
	for _, entry := range db.pricedb.Range(nil, nil, "both", entries) {
		k, nprice := entry.Key(), entry.Value().(*Price).Clone(&ndb)
		ndb.pricedb.Insert(k, nprice)
	}
	return &ndb
}

func (db *Datastore) addBalance(commodity *Commodity) error {
	return db.de.AddBalance(commodity)
}

// directive-account

func (db *Datastore) declare(value interface{}) error {
	switch v := value.(type) {
	case *Directive:
		d := v
		switch d.dtype {
		case "account":
			db.addAlias(d.accalias, d.accname)
			err := db.addPayee(d.accpayee, d.accname)
			if err != nil {
				return err
			}
			account := db.GetAccount(d.accname).(*Account)
			if len(d.acctypes) > 0 {
				account.types = d.addAccounttype(d.acctypes, account.types)
			}
			account.addNote(d.note)
			account.addAlias(d.accalias)
			account.addPayee(d.accpayee)
			account.addComments(d.comments...)
			if d.ndefault {
				db.setBalancingaccount(account.name)
			}
			db.dclrdacc = append(db.dclrdacc, d.accname)

		case "commodity":
			scanner := parsec.NewScanner([]byte(d.commdfmt))
			node, _ := NewCommodity("").Yledger(db)(scanner)
			commodity := node.(*Commodity)
			if commodity.name != "" && commodity.name != d.commdname {
				x, y := commodity.name, d.commdname
				return fmt.Errorf("name mismatching %q vs %q", x, y)
			} else if commodity.name == "" {
				commodity.name = d.commdname
				commodity.noname = true
			}

			commodity.addNote(d.note)
			if d.ndefault {
				db.setDefaultcomm(commodity.name)
			}
			if d.commdnmrkt {
				commodity.nomarket = true
			}
			if d.commdcurrn {
				commodity.currency = true
			}
			// now finally update the datastore.commodity db.
			db.commodities[commodity.name] = commodity
			db.dclrdcomm = append(db.dclrdcomm, d.commdname)

		case "payee":
			payee := db.findpayee(d.dpayee)
			for _, alias := range d.dpayeealias {
				if err := payee.addAlias(alias); err != nil {
					return err
				}
			}
			for _, uuid := range d.dpayeeuuid {
				payee.addUuid(uuid)
			}
			db.dclrdpayee = append(db.dclrdpayee, d.dpayee)

		}
		return nil
	}
	panic("unreachable code")
}

//---- api.Reporter methods

func (db *Datastore) FmtBalances(
	_ api.Datastorer, trans api.Transactor, p api.Poster,
	acc api.Accounter) [][]string {

	var rows [][]string

	if len(db.Balances()) == 0 {
		return append(rows, []string{"", "", "-"})
	}

	for _, balance := range db.Balances() {
		rows = append(rows, []string{"", "", balance.String()})
	}
	if len(rows) > 0 { // last row to include date.
		lastrow := rows[len(rows)-1]
		date := trans.Date().Format("2006/Jan/02")
		lastrow[0] = date
	}
	return rows
}

func (db *Datastore) FmtDCBalances(
	_ api.Datastorer, trans api.Transactor, p api.Poster,
	acc api.Accounter) [][]string {

	var rows [][]string

	if len(db.Balances()) == 0 {
		return append(rows, []string{"", "", "-", "-", "-"})
	}

	for _, bal := range db.Balances() {
		name := bal.Name()
		dr, cr := db.Debit(name), db.Credit(name)
		cols := []string{"", "", dr.String(), cr.String(), bal.String()}
		rows = append(rows, cols)
	}
	if len(rows) > 0 { // last row to include date.
		lastrow := rows[len(rows)-1]
		date := trans.Date().Format("2006/Jan/02")
		lastrow[0] = date
	}
	return rows
}

func (db *Datastore) FmtEquity(
	_ api.Datastorer, trans api.Transactor, _ api.Poster,
	_ api.Accounter) [][]string {

	panic("not supported")
}

func (db *Datastore) FmtPassbook(
	_ api.Datastorer, trans api.Transactor, _ api.Poster,
	_ api.Accounter) [][]string {

	panic("not supported")
}

func (db *Datastore) FmtRegister(
	_ api.Datastorer, trans api.Transactor, p api.Poster,
	acc api.Accounter) [][]string {

	panic("not supported")
}

func (db *Datastore) defaultprices() {
	_ = []string{
		"P 01/01/2000 kb 1024b",
		"P 01/01/2000 mb 1024kb",
		"P 01/01/2000 gb 1024mb",
		"P 01/01/2000 tb 1024gb",
		"P 01/01/2000 pb 1024tb",

		"P 01/01/2000 m 60s",
		"P 01/01/2000 h 60m",
	}
}
