package main

import "time"

import "github.com/prataprc/goparsec"
import s "github.com/prataprc/gosettings"

type Price struct {
	date      time.Time
	commodity time.Time
	exchange  time.Time // common exchange commodity.

	// context
	year       int
	month      int
	dateformat string
	context    s.Settings
}

func NewPrice(context s.Settings) *Price {
	price := &Price{
		year:       int(context.Int64("year")),
		month:      int(context.Int64("month")),
		dateformat: context.String("dateformat"),
		context:    context,
	}
	return price
}

func (price *Price) Y() parsec.Parser {
	noder := func(nodes []parsec.ParsecNode) parsec.ParsecNode {
		return nil
	}

	// DATE
	ydate := Ydate(price.year, price.month, price.dateformat)
	// SYMBOL
	ysymbol := parsec.Token("", "PRICESYMBOL")
	// [*|!]
	yexchange := parsec.Token("", "PRICEEXCHANGE")

	y := parsec.And(noder, ydate, ysymbol, yexchange)
	return y
}
