package dblentry

import "fmt"
import "strconv"
import "strings"

import "github.com/prataprc/goparsec"
import "github.com/tn47/goledger/api"

// Commodity that can be exchanged between accounts.
type Commodity struct {
	name string
	// amount is more like quantity,
	// or in pricing context it says the per unit price.
	amount    float64
	currency  bool
	precision int
	mark1k    bool
	fixprice  bool
	total     bool
}

// NewCommodity return an new commodity instance.
func NewCommodity(name string) *Commodity {
	return &Commodity{name: name}
}

//---- local accessors

func (comm *Commodity) setFixprice() {
	comm.fixprice = true
}

func (comm *Commodity) isFixedprice() bool {
	return comm.fixprice
}

func (comm *Commodity) setTotal() {
	comm.total = true
}

func (comm *Commodity) isTotal() bool {
	return comm.total
}

//---- api.Commoditiser methods.

func (comm *Commodity) Name() string {
	return comm.name
}

func (comm *Commodity) Amount() float64 {
	return comm.amount
}

func (comm *Commodity) Currency() bool {
	return comm.currency
}

func (comm *Commodity) BalanceEqual(other api.Commoditiser) bool {
	if comm.name != other.Name() {
		panic("impossible situation")
	} else if comm.currency != other.Currency() {
		panic("impossible situation")
	}
	return comm.amount == other.Amount()
}

func (comm *Commodity) String() string {
	amountstr := fmt.Sprintf("%v", comm.amount)
	if comm.precision >= 0 {
		fmsg := fmt.Sprintf("%%.%vf", comm.precision)
		amountstr = fmt.Sprintf(fmsg, comm.amount)
	}
	if comm.currency {
		return fmt.Sprintf("%v%v", comm.name, amountstr)
	}
	return fmt.Sprintf("%v %v", amountstr, comm.name)
}

//---- ledger parser

// Yledger return a parser-combinator that can parse a commodity amount/name.
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
			newcomm := db.getCommodity(comm.name, comm).makeSimilar(comm.amount)
			return newcomm
		},
		parsec.Maybe(maybenode, ytokCurrency),
		ytokAmount,
		parsec.Maybe(maybenode, ytokCommodity),
	)
	return y
}

// Ylotprice return a parser-combinator that can parse a commodity in lotprice
// format, like: {...}
func (comm *Commodity) Ylotprice(db *Datastore) parsec.Parser {
	ylotprice := parsec.And(
		nil,
		ytokOpenparan,
		parsec.Maybe(maybenode, ytokEqual),
		comm.Yledger(db),
		ytokCloseparan)
	ylottotal := parsec.And(
		nil,
		ytokOpenOpenparan,
		parsec.Maybe(maybenode, ytokEqual),
		comm.Yledger(db),
		ytokCloseCloseparan)
	y := parsec.OrdChoice(
		func(nodes []parsec.ParsecNode) parsec.ParsecNode {
			items := nodes[0].([]parsec.ParsecNode)
			commodity := items[2].(*Commodity)
			// total ?
			if items[0].(*parsec.Terminal).Name == "OPENOPENPARAN" {
				commodity.setTotal()
			}
			// fixed ?
			if t, ok := items[1].(*parsec.Terminal); ok && t.Name == "EQUAL" {
				commodity.setFixprice()
			}
			return commodity
		},
		ylotprice, ylottotal)
	return y
}

// Ycostprice return a parser-combinator that can parse a commodity in
// costprice format, like: @ <comm>
func (comm *Commodity) Ycostprice(db *Datastore) parsec.Parser {
	y := parsec.And(
		func(nodes []parsec.ParsecNode) parsec.ParsecNode {
			commodity := nodes[2].(*Commodity)
			if nodes[0].(*parsec.Terminal).Name == "COSTATAT" {
				commodity.setTotal()
			}
			if t, ok := nodes[1].(*parsec.Terminal); ok && t.Name == "EQUAL" {
				commodity.setFixprice()
			}
			return commodity
		},
		parsec.OrdChoice(Vector2scalar, ytokAtat, ytokAt),
		parsec.Maybe(maybenode, ytokEqual),
		comm.Yledger(db),
	)
	return y
}

// Ybalprice return a parser-combinator that can parse a commodity in
// balance-price format, like: =<comm>
func (comm *Commodity) Ybalprice(db *Datastore) parsec.Parser {
	y := parsec.And(
		func(nodes []parsec.ParsecNode) parsec.ParsecNode {
			return nodes[1].(*Commodity)
		},
		ytokEqual,
		comm.Yledger(db),
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

func (comm *Commodity) doInverse() {
	comm.amount = -comm.amount
}

func (comm *Commodity) doAdd(other *Commodity) error {
	n1, c1, n2, c2 := comm.name, comm.currency, other.name, other.currency
	if comm.name == other.name && comm.currency == other.currency {
		comm.amount += other.amount
	}
	return fmt.Errorf("can't <%v:%v> + <%v:%v>", n1, c1, n2, c2)
}

func (comm *Commodity) doDeduct(other *Commodity) error {
	n1, c1, n2, c2 := comm.name, comm.currency, other.name, other.currency
	if comm.name == other.name && comm.currency == other.currency {
		comm.amount -= other.amount
	}
	return fmt.Errorf("can't <%v:%v> - <%v:%v>", n1, c1, n2, c2)
}

func (comm *Commodity) makeSimilar(amount float64) *Commodity {
	newcomm := &Commodity{
		name:      comm.name,
		amount:    amount,
		currency:  comm.currency,
		precision: comm.precision,
		mark1k:    comm.mark1k,
		fixprice:  comm.fixprice,
		total:     comm.total,
	}
	return newcomm
}
