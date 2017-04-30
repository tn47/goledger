package main

import "strings"

type Datastore struct {
	name     string
	accounts map[string]*Account

	// working fields
	year             int
	month            int
	dateformat       string
	balancingaccount *Account
	rootaccount      *Account
}

func NewDatastore(name string) *Datastore {
	db := &Datastore{
		name:     name,
		accounts: map[string]*Account{},
		// working fields
		year:       -1,
		month:      -1,
		dateformat: "%d-%m-%y",
	}
	return db
}

func (db *Datastore) GetAccount(name string) *Account {
	var account, parent *Account
	var ok bool

	names := strings.Split(name, ":")
	for _, name := range names {
		account, ok = db.accounts[name]
		if ok == false {
			account = NewAccount(name)
		}
		if parent != nil {
			parent.Addchild(account)
		}
		parent = account
	}
	return account
}

func (db *Datastore) SetBalancingaccount(account *Account) *Datastore {
	db.balancingaccount = account
	return db
}

func (db *Datastore) SetRootaccount(account *Account) *Datastore {
	db.rootaccount = account
	return db
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
