package main

import "time"

import "github.com/prataprc/goparsec"

type Price struct {
	date      time.Time
	commodity time.Time
	exchange  time.Time // common exchange commodity.

	year       int
	month      int
	dateformat string
	context    Context
}

func NewPrice(context Context) *Price {
	price := &Price{context: context}
	if year, ok := context.Int64("year"); ok {
		price.year = int(year)
	}
	if month, ok := context.Int64("month"); ok {
		price.month = int(month)
	}
	if dateformat, ok := context.String("dateformat"); ok {
		price.dateformat = dateformat
	}
	return price
}

func (price *Price) Y() parsec.Parser {
	// P
	yp := parsec.Token("P ", "PRICESTART")
	// DATE
	ydate := Ydate(price.year, price.month, price.dateformat)
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
