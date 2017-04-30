package main

import "sort"

import "github.com/prataprc/goparsec"

var inclusives = []string{
	"asset", "liability", "capital", "equity", "income", "expense",
}

type Account struct {
	name     string
	virtual  bool
	balanced bool
	children map[string]*Account
	// from account directive
	note   string
	alias  []string
	payee  []string
	check  string
	assert string
	eval   string
}

func NewAccount(name string) *Account {
	acc := &Account{
		name: name, children: make(map[string]*Account),
		alias: []string{}, payee: []string{},
	}
	return acc
}

func (acc *Account) Y() parsec.Parser {
	y := parsec.OrdChoice(
		func(nodes []parsec.ParsecNode) parsec.ParsecNode {
			t := nodes[0].(*parsec.Terminal)
			name := string(t.Value)
			switch t.Name {
			case "TRANSACCOUNT":
				acc.name = name
			case "TRANSVACCOUNT":
				acc.virtual = true
				acc.name = name[1 : len(name)-1]
			case "TRANSBACCOUNT":
				acc.balanced = true
				acc.name = name[1 : len(name)-1]
			}
			panic("unreachable code")
		},
		parsec.Token("[a-zA-Z][a-zA-Z: ~.,;?/-]*", "TRANSACCOUNT"),
		parsec.Token(`\([a-zA-Z][a-zA-Z: ~.,;?/-]*\)`, "TRANSVACCOUNT"),
		parsec.Token(`\[[a-zA-Z][a-zA-Z: ~.,;?/-]*\]`, "TRANSBACCOUNT"),
	)
	return y
}

func (acc *Account) Addchild(child *Account) {
	if _, ok := acc.children[child.name]; ok {
		return
	}
	acc.children[child.name] = child
}

func (acc *Account) Children() []*Account {
	names := []string{}
	for _, child := range acc.children {
		names = append(names, child.name)
	}
	sort.Strings(names)
	accounts := []*Account{}
	for _, name := range names {
		accounts = append(accounts, acc.children[name])
	}
	return accounts
}

func (acc *Account) Name() string {
	return acc.name
}

func (acc *Account) SetNote(note string) *Account {
	acc.note = note
	return acc
}

func (acc *Account) SetAlias(alias string) *Account {
	acc.alias = append(acc.alias, alias)
	return acc
}

func (acc *Account) SetPayee(payee string) *Account {
	acc.payee = append(acc.payee, payee)
	return acc
}

func (acc *Account) SetCheck(check string) *Account {
	acc.check = check
	return acc
}

func (acc *Account) SetAssert(assert string) *Account {
	acc.assert = assert
	return acc
}

func (acc *Account) SetEval(eval string) *Account {
	acc.eval = eval
	return acc
}

func (acc *Account) Note() string {
	return acc.note
}

func (acc *Account) Alias() []string {
	return acc.alias
}

func (acc *Account) Payee() []string {
	return acc.payee
}

func (acc *Account) Check() string {
	return acc.check
}

func (acc *Account) Assert() string {
	return acc.assert
}

func (acc *Account) Eval() string {
	return acc.eval
}
