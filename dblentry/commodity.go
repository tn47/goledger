package dblentry

import "fmt"
import "strconv"
import "strings"

import "github.com/prataprc/goparsec"
import "github.com/tn47/goledger/api"

// Commodity that can be exchanged between accounts.
type Commodity struct {
	name  string
	notes []string
	// amount is more like quantity,
	// or in pricing context it says the per unit price.
	amount    float64
	currency  bool
	precision int
	mark1k    bool
	fixprice  bool
	total     bool
	nomarket  bool
}

// NewCommodity return an new commodity instance.
func NewCommodity(name string) *Commodity {
	return &Commodity{name: name, notes: make([]string, 0)}
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

func (comm *Commodity) addNote(note string) {
	comm.notes = append(comm.notes, note)
}

//---- api.Commoditiser methods.

func (comm *Commodity) Name() string {
	return comm.name
}

func (comm *Commodity) Notes() []string {
	return comm.notes
}

func (comm *Commodity) Amount() float64 {
	return comm.amount
}

func (comm *Commodity) Currency() bool {
	return comm.currency
}

func (comm *Commodity) ApplyAmount(other api.Commoditiser) error {
	return comm.doAdd(other.(*Commodity))
}

func (comm *Commodity) BalanceEqual(other api.Commoditiser) (bool, error) {
	if comm.name != other.Name() {
		fmsg := "mismatch in balancing commodity %q != %q"
		return false, fmt.Errorf(fmsg, comm.name, other.Name())
	} else if comm.currency != other.Currency() {
		return false, fmt.Errorf("mismatch in currency commodity")
	}
	return comm.amount == other.Amount(), nil
}

func (comm *Commodity) IsCredit() bool {
	if comm.amount < 0 {
		return true
	}
	return false
}

func (comm *Commodity) IsDebit() bool {
	if comm.amount > 0 {
		return true
	}
	return false
}
func (comm *Commodity) MakeSimilar(amount float64) api.Commoditiser {
	return comm.makeSimilar(amount)
}

func (comm *Commodity) makeSimilar(amount float64) *Commodity {
	newcomm := &Commodity{
		name:      comm.name,
		notes:     comm.notes,
		amount:    amount,
		currency:  comm.currency,
		precision: comm.precision,
		mark1k:    comm.mark1k,
		fixprice:  comm.fixprice,
		total:     comm.total,
		nomarket:  comm.nomarket,
	}
	return newcomm
}
func (comm *Commodity) String() string {
	if comm == nil {
		return ""
	}
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

func (comm *Commodity) Directive() string {
	lines := []string{fmt.Sprintf("commodity %v", comm.name)}
	for _, note := range comm.notes {
		lines = append(lines, fmt.Sprintf("    note  %v", note))
	}
	if comm.amount > 0 {
		lines = append(lines, fmt.Sprintf("    format %v", comm.String()))
	}
	if comm.nomarket {
		lines = append(lines, fmt.Sprintf("    nomarket"))
	}
	return strings.Join(lines, "\n")
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
			return comm
		},
		parsec.Maybe(maybenode, ytokCurrency),
		ytokAmount,
		parsec.Maybe(maybenode, ytokCommodity),
	)
	return y
}

// Yname return a parser-combinator that can parse a commodity name.
func (comm *Commodity) Yname(db *Datastore) parsec.Parser {
	return parsec.OrdChoice(Vector2scalar, ytokCurrency, ytokCommodity)
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

func (comm *Commodity) Clone(ndb *Datastore) *Commodity {
	ncomm := *comm
	return &ncomm
}

func (comm *Commodity) doInverse() {
	comm.amount = -comm.amount
}

func (comm *Commodity) doAdd(other *Commodity) error {
	n1, c1, n2, c2 := comm.name, comm.currency, other.name, other.currency
	if comm.name == other.name && comm.currency == other.currency {
		comm.amount += other.amount
		return nil
	}
	return fmt.Errorf("can't <%v:%v> + <%v:%v>", n1, c1, n2, c2)
}
