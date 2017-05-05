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
	equival   map[string]float64
}

func NewCommodity(name string) *Commodity {
	return &Commodity{name: name}
}

func (comm *Commodity) Similar(amount float64) *Commodity {
	newcomm := &Commodity{
		name:      comm.name,
		amount:    amount,
		currency:  comm.currency,
		precision: comm.precision,
		mark1k:    comm.mark1k,
	}
	return newcomm
}

func (comm *Commodity) String() string {
	amountstr := fmt.Sprintf("%v", comm.amount)
	if comm.precision > 0 {
		fmsg := fmt.Sprintf("%%.%vf", comm.precision)
		amountstr = fmt.Sprintf(fmsg, comm.amount)
	}
	if comm.currency {
		return fmt.Sprintf("%v%v", comm.name, amountstr)
	}
	return fmt.Sprintf("%v %v", amountstr, comm.name)
}

//---- accessors

func (comm *Commodity) Amount() float64 {
	return comm.amount
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
					fmt.Println("commodity", string(t.Value))
					comm.name, comm.currency = string(t.Value), false
				}
			}
			newcomm := db.GetCommodity(comm.name, comm).Similar(comm.amount)
			return newcomm
		},
		parsec.Maybe(maybenode, ytok_currency),
		ytok_amount,
		parsec.Maybe(maybenode, ytok_commodity),
	)
	return y
}

//---- engine

func (comm *Commodity) Firstpass(
	db *Datastore, trans *Transaction, p *Posting) error {

	return nil
}

func (comm *Commodity) Secondpass(
	db *Datastore, trans *Transaction, p *Posting) error {

	return nil
}
