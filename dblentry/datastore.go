package dblentry

import "fmt"
import "strings"

import "github.com/prataprc/goledger/api"
import "github.com/prataprc/golog"

type Datastore struct {
	name     string
	reporter api.Reporter
	transdb  *DB
	pricedb  *DB
	accntdb  map[string]*Account // full account-name -> account
	balance  float64
	// directive fields
	year         int               // year
	month        int               // month
	dateformat   string            // dataformat
	aliases      map[string]string // alias, account-alias
	payees       map[string]string // account-payee map[regex]->accountname
	rootaccount  string            // apply-account
	blncingaccnt string            // account
}

func NewDatastore(name string, reporter api.Reporter) *Datastore {
	db := &Datastore{
		name:     name,
		reporter: reporter,
		transdb:  NewDB(fmt.Sprintf("%v-transactions", name)),
		pricedb:  NewDB(fmt.Sprintf("%v-pricedb", name)),
		accntdb:  map[string]*Account{},
		// directives
		year:       -1,
		month:      -1,
		dateformat: "%Y/%m/%d %h:%n:%s", // TODO: no magic string
		aliases:    map[string]string{},
	}
	db.defaultprices()
	return db
}

//---- accessor

func (db *Datastore) GetAccount(name string) *Account {
	account, ok := db.accntdb[name]
	if ok == false {
		account = NewAccount(name)
	}
	db.accntdb[name] = account
	return account
}

func (db *Datastore) Accountnames() []string {
	accnames := []string{}
	for name := range db.accntdb {
		accnames = append(accnames, name)
	}
	return accnames
}

func (db *Datastore) HasAccount(name string) bool {
	_, ok := db.accntdb[name]
	return ok
}

func (db *Datastore) Balance() float64 {
	return db.balance
}

func (db *Datastore) SubAccounts(parentname string) []*Account {
	accounts := []*Account{}
	for name, account := range db.accntdb {
		if strings.HasPrefix(parentname, name) {
			accounts = append(accounts, account)
		}
	}
	return accounts
}

//---- engine

func (db *Datastore) Apply(obj interface{}) error {
	if trans, ok := obj.(*Transaction); ok {
		if trans.ShouldBalance() {
			var defaccount *Account
			if db.blncingaccnt != "" {
				defaccount = db.GetAccount(db.blncingaccnt)
			}
			if ok, err := trans.Autobalance1(defaccount); err != nil {
				return err
			} else if ok == false {
				return fmt.Errorf("unbalanced transaction")
			}
			log.Debugf("transaction balanced\n")
		}
		db.transdb.Insert(trans.date, trans)
		trans.Apply(db)
		return nil

	} else if price, ok := obj.(*Price); ok {
		return db.pricedb.Insert(price.when, price)

	} else if directive, ok := obj.(*Directive); ok {
		switch directive.dtype {
		case "year":
			db.SetYear(directive.year)
		case "month":
			db.SetMonth(directive.month)
		case "dateformat":
			db.SetDateformat(directive.dateformat)
		case "account":
			db.Declare(directive.account) // NOTE: this is redundant
		case "apply":
			if db.rootaccount != "" {
				fmsg := "previous `apply` directive(%v) not closed"
				return fmt.Errorf(fmsg, db.rootaccount)
			}
			db.rootaccount = directive.account.name
		case "alias":
			db.AddAlias(directive.aliasname, directive.account.name)
		case "assert":
			return fmt.Errorf("directive not-implemented")
		case "end":
			if db.rootaccount == "" {
				return fmt.Errorf("dangling `end` directive")
			}
			db.rootaccount = ""
		default:
			panic("unreachable code")
		}

	} else {
		panic("unreachable code")
	}
	return nil
}

// directive-year

func (db *Datastore) SetYear(year int) *Datastore {
	db.year = year
	return db
}

func (db *Datastore) Year() int {
	return db.year
}

// directive-month

func (db *Datastore) SetMonth(month int) *Datastore {
	db.month = month
	return db
}

func (db *Datastore) Month() int {
	return db.month
}

// directive-dateformat

func (db *Datastore) SetDateformat(format string) *Datastore {
	db.dateformat = format
	return db
}

func (db *Datastore) Dateformat() string {
	return db.dateformat
}

// directive-alias

func (db *Datastore) AddAlias(aliasname, accountname string) *Datastore {
	db.aliases[aliasname] = accountname
	return db
}

func (db *Datastore) GetAlias(aliasname string) (accountname string, ok bool) {
	accountname, ok = db.aliases[aliasname]
	return accountname, ok
}

// directive-apply-account

func (db *Datastore) SetRootaccount(name string) *Datastore {
	db.rootaccount = name
	return db
}

func (db *Datastore) Rootaccount() string {
	return db.rootaccount
}

// directive-account

func (db *Datastore) Declare(value interface{}) {
	switch v := value.(type) {
	case *Account:
		account := db.GetAccount(v.name)
		account.SetDirective(v)
		if v.defblns {
			db.SetBalancingaccount(v.name)
		}
		return

	default:
		panic("unreachable code")
	}
	panic("unreachable code")
}

func (db *Datastore) AddPayee(regex, accountname string) *Datastore {
	db.payees[regex] = accountname
	return db
}

func (db *Datastore) SetBalancingaccount(name string) *Datastore {
	db.blncingaccnt = name
	return db
}

func (db *Datastore) LookupAlias(name string) string {
	if accountname, ok := db.aliases[name]; ok {
		return accountname
	}
	return name
}

func (db *Datastore) Applyroot(name string) string {
	if db.rootaccount != "" {
		return db.rootaccount + ":" + name
	}
	return name
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
