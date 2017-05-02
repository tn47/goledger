package dblentry

import "github.com/prataprc/goparsec"

var inclusives = []string{
	"asset", "liability", "capital", "equity", "income", "expense",
}

type Account struct {
	name     string
	virtual  bool
	balanced bool
	// from account directive
	note    string
	check   string
	assert  string
	eval    string
	defblns bool
}

func NewAccount(name string) *Account {
	acc := &Account{name: name}
	return acc
}

func (tempacc *Account) Yledger(db *Datastore) parsec.Parser {
	y := parsec.OrdChoice(
		func(nodes []parsec.ParsecNode) parsec.ParsecNode {
			t := nodes[0].(*parsec.Terminal)
			name := string(t.Value)
			switch t.Name {
			case "TRANSACCOUNT":
				tempacc.name = name
				return tempacc
			case "TRANSVACCOUNT":
				tempacc.name = name[1 : len(name)-1]
				tempacc.virtual = true
				return tempacc
			case "TRANSBACCOUNT":
				tempacc.name = name[1 : len(name)-1]
				tempacc.balanced = true
				return tempacc
			}
			panic("unreachable code")
		},
		ytok_accname, ytok_vaccname, ytok_baccname,
	)
	return y
}

func (acc *Account) SetDirective(account *Account) *Account {
	acc.note = account.note
	acc.check = account.check
	acc.assert = account.assert
	acc.eval = account.eval
	return acc
}
