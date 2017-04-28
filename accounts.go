package main

import "github.com/prataprc/goparsec"

var inclusives = []string{
	"asset", "liability", "capital", "equity", "income", "expense",
}

type Account struct {
	name     string
	virtual  bool
	balanced bool

	db Datastore // read-only copy
}

func NewAccount(name string, db Datastore) *Account {
	return &Account{name: name, db: db}
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
