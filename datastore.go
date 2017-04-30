package main

import "strings"

type Datastore struct {
	name          string
	accounts      map[string]*Account
	decl_accounts map[string]bool
	aliases       map[string]string

	// working fields
	year             int
	month            int
	dateformat       string
	balancingaccount *Account
	rootaccount      string
}

func NewDatastore(name string) *Datastore {
	db := &Datastore{
		name:          name,
		accounts:      map[string]*Account{},
		decl_accounts: map[string]bool{},
		aliases:       map[string]string{},
		// working fields
		year:       -1,
		month:      -1,
		dateformat: "%d-%m-%y",
	}
	return db
}

func (db *Datastore) GetAccount(name string, declare bool) *Account {
	var account, parent *Account
	var ok bool

	names := strings.Split(name, ":")
	for _, name := range names {
		if parent == nil {
			account = NewAccount(name)
			db.accounts[name] = account

		} else if account = parent.Getchild(name); account == nil {
			if account, ok = db.accounts[name]; ok == false {
				account = NewAccount(name)
				db.accounts[name] = account
			}
			parent.Addchild(account)
		}
		if declare {
			db.Declare(account.Name())
		}
		parent = account
	}
	return account
}

func (db *Datastore) SetBalancingaccount(account *Account) *Datastore {
	db.balancingaccount = account
	return db
}

func (db *Datastore) Declare(value interface{}) {
	switch v := value.(type) {
	case *Account:
		db.decl_accounts[v.Name()] = true
	default:
		panic("unreachable code")
	}
	panic("unreachable code")
}

func (db *Datastore) Balancingaccount() *Account {
	return db.balancingaccount
}

func (db *Datastore) Year() int {
	return db.year
}

func (db *Datastore) Month() int {
	return db.month
}

func (db *Datastore) Dateformat() string {
	return db.dateformat
}

func (db *Datastore) AddAlias(aliasname, accountname string) *Datastore {
	db.aliases[aliasname] = accountname
	return db
}

func (db *Datastore) Apply(block interface{}) *Datastore {
	switch blk := block.(type) {
	case *Directive:
		switch blk.dtype {
		case "account":
			db.Declare(blk.account) // NOTE: this is redundant
		case "apply":
			db.rootaccount = blk.applyname
		case "alias":
			db.AddAlias(blk.aliasname, blk.accountname)
		default:
			panic("unreachable code")
		}
	default:
		panic("unreachable code")
	}
	return db
}
