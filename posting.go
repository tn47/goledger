package main

import "github.com/prataprc/goparsec"

type Posting struct {
	direction string // "source", target"
	virtual   bool
	balanced  bool
	account   *Account
	amount    *Amount
	note      string
}

func NewPosting() *Posting {
	return &Posting{}
}

func (p *Posting) Y(db *Datastore) parsec.Parser {
	account := NewAccount("")
	yposting := parsec.And(nil, account.Y(), ytok_postamount, ytok_postnote)

	y := parsec.OrdChoice(
		func(nodes []parsec.ParsecNode) parsec.ParsecNode {
			if len(nodes) == 3 {
				p.account = nodes[0].(*Account)
				p.amount = NewAmount(string(nodes[1].(*parsec.Terminal).Value))
				p.note = string(nodes[2].(*parsec.Terminal).Value)
				return p

			} else if len(nodes) == 1 {
				return Transnote(string(nodes[0].(*parsec.Terminal).Value))
			}
			panic("unreachable code")
		},
		yposting, ytok_persnote,
	)
	return y
}
