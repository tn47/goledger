package main

import "time"

import "github.com/prataprc/goparsec"

type Price struct {
	date      time.Time
	commodity time.Time
	exchange  time.Time // common exchange commodity.
}

func NewPrice() *Price {
	price := &Price{}
	return price
}

func (price *Price) Y(db *Datastore) parsec.Parser {
	// P
	yp := parsec.Token("P ", "PRICESTART")
	// DATE
	ydate := Ydate(db.Year(), db.Month(), db.Dateformat())
	// SYMBOL
	ysymbol := parsec.Token("", "PRICESYMBOL")
	// [*|!]
	yexchange := parsec.Token("", "PRICEEXCHANGE")

	y := parsec.And(
		func(nodes []parsec.ParsecNode) parsec.ParsecNode {
			return price
		},
		yp, ydate, ysymbol, yexchange,
	)
	return y
}
