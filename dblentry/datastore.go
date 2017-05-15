package dblentry

import "fmt"
import "sort"
import "time"

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
	name     string
	reporter api.Reporter

	transdb     *DB
	pricedb     *DB
	accntdb     map[string]*Account // full account-name -> account
	commodities map[string]*Commodity

	defaultcomm string
	comments    []string
	pass        ParsePhase
	balance     map[string]*Commodity

	// directive fields
	currdate     time.Time
	rootaccount  string            // apply-account
	blncingaccnt string            // account
	aliases      map[string]string // alias, account-alias
	payees       map[string]string // account-payee map[regex]->accountname
}

// NewDatastore return a new datastore.
func NewDatastore(name string, reporter api.Reporter) *Datastore {
	db := &Datastore{
		name:     name,
		reporter: reporter,

		transdb:     NewDB(fmt.Sprintf("%v-transactions", name)),
		pricedb:     NewDB(fmt.Sprintf("%v-pricedb", name)),
		accntdb:     map[string]*Account{},
		commodities: map[string]*Commodity{},

		comments: []string{},
		balance:  make(map[string]*Commodity),
		pass:     DBSTART,
		// directives
		currdate: time.Now(),
		aliases:  map[string]string{},
	}
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
	if name == "" && db.defaultcomm != "" {
		return db.commodities[db.defaultcomm]
	}
	if db.defaultcomm == "" && name != "" {
		db.defaultcomm = name
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

func (db *Datastore) setYear(year int) *Datastore {
	db.currdate = time.Date(year, 1, 1, 0, 0, 0, 0, time.Local)
	return db
}

func (db *Datastore) getYear() int {
	return db.currdate.Year()
}

func (db *Datastore) setCurrentDate(date time.Time) *Datastore {
	db.currdate = date
	return db
}

func (db *Datastore) currentDate() time.Time {
	return db.currdate
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
		db.comments = append(db.comments, comment.line)
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

// directive-alias

func (db *Datastore) addAlias(aliasname, accountname string) *Datastore {
	db.aliases[aliasname] = accountname
	return db
}

func (db *Datastore) getAlias(aliasname string) (accountname string, ok bool) {
	accountname, ok = db.aliases[aliasname]
	return accountname, ok
}

// directive-account

func (db *Datastore) declare(value interface{}) error {
	switch v := value.(type) {
	case *Account:
		account := db.GetAccount(v.name).(*Account)
		account.setDirective(v)
		if v.defblns {
			db.setBalancingaccount(v.name)
		}
		return nil
	}
	panic("unreachable code")
}

func (db *Datastore) addPayee(regex, accountname string) *Datastore {
	db.payees[regex] = accountname
	return db
}

func (db *Datastore) setBalancingaccount(name string) *Datastore {
	db.blncingaccnt = name
	return db
}

func (db *Datastore) lookupAlias(name string) string {
	if accountname, ok := db.aliases[name]; ok {
		return accountname
	}
	return name
}

func (db *Datastore) applyroot(name string) string {
	if db.rootaccount != "" {
		return db.rootaccount + ":" + name
	}
	return name
}

//---- api.Reporter methods

func (db *Datastore) FmtBalances(
	_ api.Datastorer, trans api.Transactor, p api.Poster,
	acc api.Accounter) [][]string {

	var rows [][]string

	if len(db.Balances()) == 0 {
		return append(rows, []string{"", "", "-"})
	}

	balances := db.Balances()
	for _, balance := range balances[:len(balances)-1] {
		rows = append(rows, []string{"", "", balance.String()})
	}
	balance := balances[len(balances)-1]
	date := trans.Date().Format("2006/Jan/02")
	rows = append(rows, []string{date, "", balance.String()})
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
