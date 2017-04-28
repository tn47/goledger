package main

import "github.com/prataprc/goparsec"
import s "github.com/prataprc/gosettings"

type DirectiveAccount struct {
	name   string
	note   string
	alias  string
	payee  string
	check  string
	assert string
	eval   string

	context s.Settings
}

func NewDirectiveAccount(context s.Settings) *DirectiveAccount {
	return &DirectiveAccount{context: context}
}

func (p *DirectiveAccount) Y() parsec.Parser {
	poster := func(nodes []parsec.ParsecNode) parsec.ParsecNode {
		return nil
	}

	accounter := func(nodes []parsec.ParsecNode) parsec.ParsecNode {
		t := nodes[0].(*parsec.Terminal)
		switch t.Name {
		case "TRANSACCOUNT", "TRANSVACCOUNT", "TRANSBACCOUNT":
			return NewAccount(string(t.Value))
		}
		panic("unreachable code")
	}

	// ACCOUNT
	yaccount := parsec.OrdChoice(
		accounter,
		parsec.Token("[a-zA-Z][a-zA-Z: ~.,;?/-]*", "TRANSACCOUNT"),
		parsec.Token(`\([a-zA-Z][a-zA-Z: ~.,;?/-]*\)`, "TRANSVACCOUNT"),
		parsec.Token(`\[[a-zA-Z][a-zA-Z: ~.,;?/-]*\]`, "TRANSBACCOUNT"),
	)
	// AMOUNT
	yamount := parsec.Token("[^;]+", "TRANSAMOUNT")
	// [; NOTE]
	ynote := parsec.Token(";[^;]+", "TRANSNOTE")

	yposting := parsec.And(nil, yaccount, yamount, ynote)
	ypersnote := parsec.Token(";[^;]+", "TRANSPNOTE")

	y := parsec.OrdChoice(poster, yposting, ypersnote)
	return y
}
