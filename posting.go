package main

import "github.com/prataprc/goparsec"

type Posting struct {
	direction string // "source", target"
	virtual   bool
	balanced  bool
	account   *Account
	amount    *Amount
	note      *Note
}

func NewPosting() *Posting {
	return &Posting{}
}

func (p *Posting) Y() parsec.Parser {
	// ACCOUNT
	yaccount := parsec.OrdChoice(
		func(nodes []parsec.ParsecNode) parsec.ParsecNode {
			t := nodes[0].(*parsec.Terminal)
			switch t.Name {
			case "TRANSACCOUNT", "TRANSVACCOUNT", "TRANSBACCOUNT":
				return NewAccount(string(t.Value))
			}
			panic("unreachable code")
		},
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

	y := parsec.OrdChoice(
		func(nodes []parsec.ParsecNode) parsec.ParsecNode {
			return p
		},
		yposting, ypersnote,
	)
	return y
}
