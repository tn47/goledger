package dblentry

import "fmt"
import "time"
import "strings"

import "github.com/prataprc/goparsec"
import "github.com/prataprc/golog"
import "github.com/prataprc/goledger/api"

const (
	PostUncleared = "uncleared"
	PostCleared   = "cleared"
	PostPending   = "pending"
)

var prefix2state = map[rune]string{
	'*': PostCleared,
	'!': PostPending,
}

type Posting struct {
	trans     *Transaction
	account   *Account
	commodity *Commodity

	lotprice  *Commodity
	lotdate   time.Time
	costprice *Commodity
	balprice  *Commodity

	tags     []string
	metadata map[string]interface{}
	note     string
}

func NewPosting(trans *Transaction) *Posting {
	return &Posting{
		trans:    trans,
		tags:     []string{},
		metadata: map[string]interface{}{},
	}
}

//---- accessor

func (p *Posting) Commodity() api.Commoditiser {
	return p.commodity
}

func (p *Posting) Account() api.Accounter {
	return p.account
}

func (p *Posting) Metadata(key string) interface{} {
	if value, ok := p.metadata[key]; ok {
		return value
	}
	return p.trans.Metadata(key)
}

func (p *Posting) SetMetadata(key string, value interface{}) {
	p.metadata[key] = value
}

func (p *Posting) Payee() string {
	payee := p.Metadata("payee")
	if payee == nil {
		return ""
	}
	return payee.(string)
}

func (p *Posting) State() string {
	state := p.Metadata("state")
	if state == nil {
		return ""
	}
	return state.(string)
}

//---- ledger parser

func (p *Posting) Yledger(db *Datastore) parsec.Parser {
	account := NewAccount("")
	comm := NewCommodity("")
	lotprice := NewCommodity("")
	costprice := NewCommodity("")
	balprice := NewCommodity("")

	ylotdate := parsec.And(nil, ytok_opensqrt, Ydate(db.Year()), ytok_closesqrt)

	yposting := parsec.And(
		nil,
		parsec.Maybe(maybenode, ytok_prefix),
		account.Ypostaccn(db),
		parsec.Maybe(maybenode, comm.Yledger(db)),
		parsec.Maybe(maybenode, lotprice.Ylotprice(db)),
		parsec.Maybe(maybenode, ylotdate),
		parsec.Maybe(maybenode, costprice.Ycostprice(db)),
		parsec.Maybe(maybenode, balprice.Ybalprice(db)),
		parsec.Maybe(maybenode, ytok_postnote),
	)

	y := parsec.OrdChoice(
		func(nodes []parsec.ParsecNode) parsec.ParsecNode {
			switch items := nodes[0].(type) {
			case []parsec.ParsecNode:
				// prefix
				if t, ok := items[0].(*parsec.Terminal); ok {
					p.SetMetadata("state", prefix2state[[]rune(t.Value)[0]])
				}

				p.account = p.fixaccount(db, items[1])     // account
				p.commodity = p.fixcommodity(db, items[2]) // commodity
				p.lotprice = p.fixlotprice(items[3])       // lot price
				p.lotdate = p.fixlotdate(items[4])         // lot date
				p.costprice = p.fixcostprice(items[5])     // cost price
				p.balprice = p.fixbalprice(items[6])       // balance price

				if p.lotprice != nil && lotprice.currency == false {
					return fmt.Errorf("lot price must be currency")
				} else if p.costprice != nil && costprice.currency == false {
					return fmt.Errorf("cost price must be currency")
				}
				if x, y := p.balprice, p.commodity; x != nil && y != nil {
					if x.name != y.name {
						fmsg := "balance-commodity(%v) != posting-commodity(%v)"
						return fmt.Errorf(fmsg, x.name, y.name)
					}
				}

				// optionally tags or tagkv or note
				if note, ok := items[7].(*parsec.Terminal); ok {
					scanner := parsec.NewScanner([]byte(note.Value))
					if node, _ := NewTag().Yledger(db)(scanner); node == nil {
						p.note = string(note.Value)

					} else {
						tag := node.(*Tags)
						p.tags = append(p.tags, tag.tags...)
						for k, v := range tag.tagm {
							p.SetMetadata(k, v)
						}
					}
				}

				fmsg := "posting.yledger account:%v commodity:%v\n"
				log.Debugf(fmsg, p.account, p.commodity)
				return p

			case *parsec.Terminal:
				inp := []byte(strings.TrimLeft(items.Value, ";"))
				scanner := parsec.NewScanner(inp)
				node, _ := NewTag().Yledger(db)(scanner)
				if node == nil {
					log.Debugf("posting.yledger %v\n", string(items.Value))
					return Transnote(items.Value)
				}
				log.Debugf("posting.yledger %v\n", node)
				return node.(*Tags)
			}
			fmsg := "unreachable code posting: len(nodes): %v"
			panic(fmt.Errorf(fmsg, len(nodes)))
		},
		yposting,
		ytok_transnote,
	)
	return y
}

func (p *Posting) fixaccount(db *Datastore, item interface{}) *Account {
	account := item.(*Account)
	accname := db.Applyroot(db.LookupAlias(account.name))
	paccount := db.GetAccount(accname).(*Account)
	paccount.virtual = account.virtual
	paccount.balanced = account.balanced
	return paccount
}

func (p *Posting) fixcommodity(db *Datastore, item interface{}) *Commodity {
	if commodity, ok := item.(*Commodity); ok {
		return db.GetCommodity(
			commodity.name, commodity,
		).Similar(commodity.amount)
	}
	return nil
}

func (p *Posting) fixlotprice(item interface{}) *Commodity {
	if lotprice, ok := item.(*Commodity); ok {
		return lotprice
	}
	return nil
}

func (p *Posting) fixlotdate(item interface{}) (tm time.Time) {
	if lotnodes, ok := item.([]parsec.ParsecNode); ok {
		return lotnodes[1].(time.Time)
	}
	return
}

func (p *Posting) fixcostprice(item interface{}) *Commodity {
	if costprice, ok := item.(*Commodity); ok {
		return costprice
	}
	return nil
}

func (p *Posting) fixbalprice(item interface{}) *Commodity {
	if balprice, ok := item.(*Commodity); ok {
		return balprice
	}
	return nil
}

//---- engine

func (p *Posting) TryAtPrice() *Commodity {
	if p.costprice != nil && p.commodity.currency == false {
		return p.costprice.Similar(p.commodity.amount * p.costprice.amount)
	}
	return p.commodity.Similar(p.commodity.amount)
}

func (p *Posting) Firstpass(db *Datastore, trans *Transaction) error {
	if err := p.account.Firstpass(db, trans, p); err != nil {
		return err
	}
	if err := p.commodity.Firstpass(db, trans, p); err != nil {
		return err
	}
	return nil
}

func (p *Posting) Secondpass(db *Datastore, trans *Transaction) error {

	db.AddBalance(p.commodity)
	p.account.SetPosting()

	if err := p.account.Secondpass(db, trans, p); err != nil {
		return err
	}
	if err := p.commodity.Secondpass(db, trans, p); err != nil {
		return err
	}

	return nil
}

//---- Reporting

func (p *Posting) FmtBalances(
	db api.Datastorer, trans api.Transactor, _ api.Poster,
	_ api.Accounter) [][]string {

	panic("not supported")
}

func (p *Posting) FmtRegister(
	db api.Datastorer, trans api.Transactor, _ api.Poster,
	_ api.Accounter) [][]string {

	panic("not supported")
}
