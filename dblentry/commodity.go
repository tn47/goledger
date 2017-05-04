package dblentry

import "fmt"
import "strconv"
import "strings"

import "github.com/prataprc/goparsec"

type Commodity struct {
	name      string
	amount    float64
	currency  bool
	precision int
	mark1k    bool
}

func NewCommodity() *Commodity {
	return &Commodity{}
}

//---- accessors

func (comm *Commodity) String() string {
	if comm.currency {
		return fmt.Sprintf("%v %v", comm.name, comm.amount)
	}
	return fmt.Sprintf("Commodity<%v %q>", comm.amount, comm.name)
}

//---- ledger parser

func (comm *Commodity) Yledger(db *Datastore) parsec.Parser {
	parseprecision := func(amount string) int {
		parts := strings.Split(amount, ".")
		if len(parts) == 2 {
			return len(parts[1])
		}
		return 0
	}

	y := parsec.And(
		func(nodes []parsec.ParsecNode) parsec.ParsecNode {
			for _, node := range nodes {
				t, ok := node.(*parsec.Terminal)
				if ok == false {
					continue
				}
				var err error
				switch t.Name {
				case "CURRENCY":
					comm.name, comm.currency = string(t.Value), true
				case "AMOUNT":
					comm.mark1k = strings.Contains(string(t.Value), ",")
					amount := strings.Replace(string(t.Value), ",", "", -1)
					comm.precision = parseprecision(amount)
					comm.amount, err = strconv.ParseFloat(amount, 64)
					if err != nil {
						panic(err)
					}

				case "COMMODITY":
					comm.name, comm.currency = string(t.Value), false
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

//---- engine

func (comm *Commodity) Balance(tmpl *Commodity, amount float64) *Commodity {
	comm.name, comm.currency = tmpl.name, tmpl.currency
	comm.precision, comm.mark1k = tmpl.precision, tmpl.mark1k
	comm.amount = amount
	return comm
}
