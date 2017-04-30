package main

import "strconv"
import "strings"

import "github.com/prataprc/goparsec"

type Commodity struct {
	currency  string
	amount    float64
	symbol    string
	precision int
	mark1k    bool
}

func NewCommodity() *Commodity {
	return &Commodity{}
}

func (comm *Commodity) Y(db *Datastore) parsec.Parser {
	y := parsec.And(
		func(nodes []parsec.ParsecNode) parsec.ParsecNode {
			for _, node := range nodes {
				var err error

				t := node.(*parsec.Terminal)
				switch t.Name {
				case "CURRENCY":
					comm.currency = string(t.Value)
				case "AMOUNT":
					amount := strings.Replace(string(t.Value), ",", "", -1)
					comm.precision = comm.parseprecision(amount)
					comm.amount, err = strconv.ParseFloat(amount, 64)
					if err != nil {
						panic(err)
					}

				case "COMMODITY":
					comm.symbol = string(t.Value)
				}
			}
			return comm
		},
		parsec.Maybe(maybenode, ytok_currency),
		ytok_amount,
		parsec.Maybe(maybenode, ytok_commodity),
	)
	return y
}

func (comm *Commodity) parseprecision(amount string) int {
	parts := strings.Split(amount, ".")
	if len(parts) == 2 {
		return len(parts[1])
	}
	return 0
}
