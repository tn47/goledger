package dblentry

import "fmt"
import "sort"

import "github.com/tn47/goledger/api"
import "github.com/prataprc/golog"

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
	firstpass

	// changes with every second pass.
	reporter    api.Reporter
	pass        ParsePhase
	commodities map[string]*Commodity
	accntdb     map[string]*Account // full account-name -> account
	balance     map[string]*Commodity
	transdb     *DB
	pricedb     *DB
}

// NewDatastore return a new datastore.
func NewDatastore(name string, reporter api.Reporter) *Datastore {
	db := &Datastore{
		name:     name,
		reporter: reporter,

		pass:        DBSTART,
		transdb:     NewDB(fmt.Sprintf("%v-transactions", name)),
		pricedb:     NewDB(fmt.Sprintf("%v-pricedb", name)),
		accntdb:     map[string]*Account{},
		commodities: map[string]*Commodity{},

		balance: make(map[string]*Commodity),
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

func (db *Datastore) HasAccount(name string) bool {
	_, ok := db.accntdb[name]
	return ok
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
	return accnames
}

func (db *Datastore) Balance(obj interface{}) (balance api.Commoditiser) {
	db.assertfirstpass()

	switch v := obj.(type) {
	case api.Commoditiser:
		balance, _ = db.balance[v.Name()]
	case string:
		balance, _ = db.balance[v]
	}
	return balance
}

func (db *Datastore) Balances() []api.Commoditiser {
	db.assertfirstpass()

	keys := []string{}
	for name := range db.balance {
		keys = append(keys, name)
	}
	sort.Strings(keys)
	comms := []api.Commoditiser{}
	for _, key := range keys {
		comms = append(comms, db.balance[key])
	}
	return comms
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
	var kvfull []KV

	for _, kv := range db.transdb.Range(nil, nil, "both", kvfull) {
		trans := kv.v.(*Transaction)
		if err := trans.Secondpass(db); err != nil {
			return err
		}
	}
	return nil
}

func (db *Datastore) Clone(nreporter api.Reporter) api.Datastorer {
	ndb := *db
	ndb.reporter = nreporter

	ndb.commodities = map[string]*Commodity{}
	for name, commodity := range db.commodities {
		ndb.commodities[name] = commodity.Clone(&ndb)
	}

	ndb.accntdb = map[string]*Account{}
	for name, account := range db.accntdb {
		ndb.accntdb[name] = account.Clone(&ndb)
	}

	ndb.balance = map[string]*Commodity{}
	for name, commodity := range db.balance {
		ndb.balance[name] = commodity.Clone(&ndb)
	}

	ndb.transdb = NewDB(fmt.Sprintf("%v-transactions", ndb.name))
	for _, kv := range db.transdb.Range(nil, nil, "both", []KV{}) {
		k, ntrans := kv.k, kv.v.(*Transaction).Clone(&ndb)
		ndb.transdb.Insert(k, ntrans)
	}

	ndb.pricedb = NewDB(fmt.Sprintf("%v-pricedb", ndb.name))
	for _, kv := range db.pricedb.Range(nil, nil, "both", []KV{}) {
		k, nprice := kv.k, kv.v.(*Price).Clone(&ndb)
		ndb.pricedb.Insert(k, nprice)
	}
	return &ndb
}

func (db *Datastore) addBalance(commodity *Commodity) {
	if balance, ok := db.balance[commodity.name]; ok {
		balance.amount += commodity.amount
		db.balance[commodity.name] = balance
		return
	}
	db.balance[commodity.name] = commodity.makeSimilar(commodity.amount)
}

func (db *Datastore) deductBalance(commodity *Commodity) {
	if balance, ok := db.balance[commodity.name]; ok {
		balance.amount -= commodity.amount
		db.balance[commodity.name] = balance
		return
	}
	db.balance[commodity.name] = commodity.makeSimilar(commodity.amount)
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
			account.addNote(d.accnote)
			account.addAlias(d.accalias)
			account.addPayee(d.accpayee)
			if d.accdefault {
				db.setBalancingaccount(account.name)
			}
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

func (db *Datastore) FmtEquity(
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
