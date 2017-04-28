package main

import "github.com/prataprc/goparsec"

type Posting struct {
	direction string // "source", target"
	virtual   bool
	balanced  bool
	account   *Account
	amount    *Amount
	note      *Note

	context Context
}

func NewPosting() *Posting {
	return &Posting{}
}

func (p *Posting) Y() parsec.Parser {
	account := NewAccount("", p.context)
	// ACCOUNT
	// AMOUNT
	yamount := parsec.Token("[^;]+", "TRANSAMOUNT")
	// [; NOTE]
	ynote := parsec.Token(";[^;]+", "TRANSNOTE")

	yposting := parsec.And(nil, account.Y(), yamount, ynote)
	ypersnote := parsec.Token(";[^;]+", "TRANSPNOTE")

	y := parsec.OrdChoice(
		func(nodes []parsec.ParsecNode) parsec.ParsecNode {
			return p
		},
		yposting, ypersnote,
	)
	return y
}
