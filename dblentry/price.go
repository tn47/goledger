package dblentry

import "time"

import "github.com/prataprc/goparsec"

// Price equivalence between commodities, TBD...
type Price struct {
	when  time.Time
	this  *Commodity
	other *Commodity
}

// NewPrice return a new Price instance.
func NewPrice() *Price {
	price := &Price{}
	return price
}

//---- ledger parser

// Yledger return a parser-combinator that can parse a price directive.
func (price *Price) Yledger(db *Datastore) parsec.Parser {
	comm := NewCommodity("")

	y := parsec.And(
		func(nodes []parsec.ParsecNode) parsec.ParsecNode {
			price.when = nodes[1].(time.Time)
			//price.this = NewCommodity("")
			//price.this.name = string(nodes[2].(*parsec.Terminal).Value)
			//price.this.amount = 1
			//price.other = nodes[3].(*Commodity)
			return price
		},
		ytokPrice,           // P
		Ydate(db.getYear()), // DATE
		ytokCommodity,       // SYMBOL
		comm.Yledger(db),
	)
	return y
}

//---- Engine

func (price *Price) Firstpass(db *Datastore) error {
	return nil
}

func (price *Price) Secondpass(db *Datastore) error {
	return nil
}
