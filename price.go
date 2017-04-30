package main

import "time"

import "github.com/prataprc/goparsec"

type Price struct {
	when time.Time
	this *Commodity
	that *Commodity
}

func NewPrice() *Price {
	price := &Price{}
	return price
}

func (price *Price) Y(db *Datastore) parsec.Parser {
	commodity := NewCommodity()

	y := parsec.And(
		func(nodes []parsec.ParsecNode) parsec.ParsecNode {
			price.when = nodes[1].(time.Time)
			price.this = NewCommodity()
			price.this.name = string(nodes[2].(*parsec.Terminal).Value)
			price.this.amount = 1
			price.that = nodes[3].(*Commodity)
			return price
		},
		ytok_price, // P
		Ydate(db.Year(), db.Month(), db.Dateformat()), // DATE
		ytok_commodity,                                // SYMBOL
		commodity.Y(db),
	)
	return y
}
