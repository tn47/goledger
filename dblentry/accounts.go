package dblentry

import "fmt"
import "sort"
import "strings"

import "github.com/prataprc/goparsec"
import "github.com/prataprc/goledger/api"

var inclusives = []string{
	"asset", "liability", "capital", "equity", "income", "expense",
}

type Account struct {
	name     string
	virtual  bool
	balanced bool
	balance  map[string]*Commodity
	// from account directive
	note      string
	aliasname string
	payee     string
	check     string
	assert    string
	eval      string
	defblns   bool
}

func NewAccount(name string) *Account {
	acc := &Account{
		name:    name,
		balance: make(map[string]*Commodity),
	}
	return acc
}

//---- accessors

func (acc *Account) SetOpeningbalance(commodity *Commodity) *Account {
	acc.balance[commodity.name] = commodity.Similar(commodity.amount)
	return acc
}

func (acc *Account) SetDirective(account *Account) *Account {
	acc.note = account.note
	acc.check = account.check
	acc.assert = account.assert
	acc.eval = account.eval
	return acc
}

func (acc *Account) Name() string {
	return acc.name
}

func (acc *Account) Balance(obj interface{}) (balance api.Commoditiser) {
	switch v := obj.(type) {
	case *Commodity:
		balance, _ = acc.balance[v.name]
	case string:
		balance, _ = acc.balance[v]
	}
	return balance
}

func (acc *Account) Balances() []api.Commoditiser {
	keys := []string{}
	for name := range acc.balance {
		keys = append(keys, name)
	}
	sort.Strings(keys)
	comms := []api.Commoditiser{}
	for _, key := range keys {
		comms = append(comms, acc.balance[key])
	}
	return comms
}

func (acc *Account) Virtual() bool {
	return acc.virtual
}

func (acc *Account) Balanced() bool {
	return acc.balanced
}

func (acc *Account) Formatedname() string {
	if acc.virtual && !acc.balanced {
		return fmt.Sprintf("(%s)", acc.name)
	} else if acc.virtual {
		return fmt.Sprintf("[%s]", acc.name)
	}
	return acc.name
}

func (acc *Account) String() string {
	return fmt.Sprintf("%v", acc.name)
}

//---- ledger parser

func (acc *Account) Yledger(db *Datastore) parsec.Parser {
	y := parsec.OrdChoice(
		func(nodes []parsec.ParsecNode) parsec.ParsecNode {
			t := nodes[0].(*parsec.Terminal)
			name := string(t.Value)
			switch t.Name {
			case "FULLACCNM":
				acc.name = name
				acc.virtual, acc.balanced = false, true
				return acc
			case "VFULLACCNM":
				acc.name = name[1 : len(name)-1]
				acc.virtual, acc.balanced = true, false
				return acc
			case "BFULLACCNM":
				acc.name = name[1 : len(name)-1]
				acc.virtual, acc.balanced = true, true
				return acc
			}
			panic(fmt.Errorf("unreachable code: terminal(%q)", t.Name))
		},
		ytok_accname, ytok_vaccname, ytok_baccname,
	)
	return y
}

//---- engine

func (acc *Account) Total(db *Datastore, trans *Transaction, p *Posting) error {
	accountnames := db.Accountnames()
	for _, name := range accountnames {
		prefix := strings.Trim(Lcp([]string{name, acc.name}), ":")
		if name == acc.name || prefix == "" {
			continue
		}
		if db.HasAccount(prefix) == false {
			consacc := db.GetAccount(prefix)
			for _, bc := range db.GetAccount(name).balance {
				consacc.AddBalance(bc)
			}
			err := db.reporter.BubblePosting(db, trans, p, consacc)
			if err != nil {
				return err
			}

			consacc = db.GetAccount(prefix)
			consacc.AddBalance(p.commodity)
			err = db.reporter.BubblePosting(db, trans, p, consacc)
			if err != nil {
				return err
			}

		} else if prefix == name {
			consacc := db.GetAccount(prefix)
			consacc.AddBalance(p.commodity)
			err := db.reporter.BubblePosting(db, trans, p, consacc)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (acc *Account) Firstpass(
	db *Datastore, trans *Transaction, p *Posting) error {

	return nil
}

func (acc *Account) Secondpass(
	db *Datastore, trans *Transaction, p *Posting) error {

	db.AddBalance(p.commodity)
	acc.AddBalance(p.commodity)
	if err := db.reporter.Posting(db, trans, p, acc); err != nil {
		return err
	}

	if err := acc.Total(db, trans, p); err != nil {
		return err
	}
	return nil
}

func (acc *Account) AddBalance(commodity *Commodity) {
	if balance, ok := acc.balance[commodity.name]; ok {
		balance.amount += commodity.amount
		acc.balance[commodity.name] = balance
		return
	}
	acc.balance[commodity.name] = commodity.Similar(commodity.amount)
}

func (acc *Account) DeductBalance(commodity *Commodity) {
	if balance, ok := acc.balance[commodity.name]; ok {
		balance.amount -= commodity.amount
		acc.balance[commodity.name] = balance
		return
	}
	acc.balance[commodity.name] = commodity.Similar(commodity.amount)
}

func FitAccountname(name string, maxwidth int) string {
	if len(name) < maxwidth {
		return name
	}
	scraplen := maxwidth - len(name)
	names := []string{}
	for _, name := range strings.Split(name, ":") {
		if scraplen <= 0 {
			names = append(names, name)
		}
		if len(name[3:]) < scraplen {
			names = append(names, name[:3])
			scraplen -= len(name[3:])
			continue
		}
		names = append(names, name[:len(name)-scraplen])
		scraplen = 0
	}
	return strings.Join(names, ":")
}
