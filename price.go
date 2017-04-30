package main

import "time"

import "github.com/prataprc/goparsec"

type Price struct {
	date      time.Time
	commodity time.Time
	exchange  time.Time // common exchange commodity.

	db *Datastore // read-only copy
}

func NewPrice(db *Datastore) *Price {
	price := &Price{db: db}
	return price
}

func (price *Price) Y() parsec.Parser {
	// P
	yp := parsec.Token("P ", "PRICESTART")
	// DATE
	ydate := Ydate(
		price.db.Year(), price.db.Month(), price.db.Dateformat(),
	)
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
