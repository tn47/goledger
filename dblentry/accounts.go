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
	name       string
	virtual    bool
	balanced   bool
	hasposting bool
	balance    map[string]*Commodity
	children   map[string]*Account
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
		name:     name,
		balance:  make(map[string]*Commodity),
		children: make(map[string]*Account, 0),
	}
	return acc
}

func (acc *Account) SetPosting() {
	acc.hasposting = true
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

func (acc *Account) HasPosting() bool {
	return acc.hasposting
}

func (acc *Account) Formatedname() string {
	if acc.virtual {
		return fmt.Sprintf("(%s)", acc.name)
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
			name := strings.Trim(string(t.Value), " \t")
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

func (acc *Account) Ypostaccn(db *Datastore) parsec.Parser {
	yacconly := parsec.And(
		func(nodes []parsec.ParsecNode) parsec.ParsecNode {
			return nodes[0]
		},
		ytok_accname, parsec.Parser(parsec.End),
	)
	y := parsec.OrdChoice(
		func(nodes []parsec.ParsecNode) parsec.ParsecNode {
			t := nodes[0].(*parsec.Terminal)
			name := strings.Trim(string(t.Value), " \t")
			switch t.Name {
			case "POSTACCN1":
				acc.name = name
				acc.virtual, acc.balanced = false, true
				return acc
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
		ytok_postacc1, ytok_vaccname, ytok_baccname, yacconly,
	)
	return y
}

//---- engine

func (acc *Account) Total(db *Datastore, trans *Transaction, p *Posting) error {
	names := SplitAccount(acc.name)
	parts := []string{}
	for _, name := range names[:len(names)-1] {
		parts = append(parts, name)
		fullname := strings.Join(parts, ":")
		consacc := db.GetAccount(fullname).(*Account)
		consacc.AddBalance(p.commodity)
		err := db.reporter.BubblePosting(db, trans, p, consacc)
		if err != nil {
			return err
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

	p.account.AddBalance(p.commodity)

	balance := p.account.Balance(p.commodity.name)
	if p.balprice != nil && balance.BalanceEqual(p.balprice) == false {
		accname := p.account.name
		fmsg := "account(%v) should balance as %s, got %s\n"
		return fmt.Errorf(fmsg, accname, p.balprice.String(), balance.String())
	}

	if err := db.reporter.Posting(db, trans, p); err != nil {
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
	for _, name := range SplitAccount(name) {
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
	return JoinAccounts(names)
}

func SplitAccount(name string) []string {
	return strings.Split(strings.Trim(name, ":"), ":")
}

func JoinAccounts(names []string) string {
	return strings.Join(names, ":")
}

//---- Reporting

func (acc *Account) FmtBalances(
	db api.Datastorer, trans api.Transactor, p api.Poster,
	_ api.Accounter) [][]string {

	if len(acc.Balances()) == 0 {
		return nil
	}

	balances, rows := acc.Balances(), make([][]string, 0)
	for _, balance := range balances[:len(balances)-1] {
		rows = append(rows, []string{"", "", balance.String()})
	}
	balance := balances[len(balances)-1]
	date := trans.Date().Format("2006/Jan/02")
	rows = append(rows, []string{date, acc.Name(), balance.String()})
	return rows
}

func (acc *Account) FmtRegister(
	db api.Datastorer, trans api.Transactor, p api.Poster,
	_ api.Accounter) [][]string {

	panic("not supported")
}
