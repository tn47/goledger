package dblentry

import "fmt"
import "time"
import "strings"

import "github.com/prataprc/goparsec"
import "github.com/prataprc/golog"
import "github.com/prataprc/goledger/api"

type Posting struct {
	trans     *Transaction
	account   *Account
	commodity *Commodity
	atprice   *Commodity
	lotprice  *Commodity
	fixprice  *Commodity
	lotdate   time.Time
	tags      []string
	metadata  map[string]interface{}
	note      string
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

func (p *Posting) Payee() string {
	payee := p.Metadata("payee")
	if payee == nil {
		return ""
	}
	return payee.(string)
}

func (p *Posting) Metadata(key string) interface{} {
	if value, ok := p.metadata[key]; ok {
		return value
	}
	return p.trans.Metadata(key)
}

//---- ledger parser

func (p *Posting) Yledger(db *Datastore) parsec.Parser {
	account := NewAccount("")
	comm := NewCommodity("")
	atprice := NewCommodity("")
	lotprice := NewCommodity("")

	ylotprice := parsec.And(
		nil,
		ytok_openparan,
		parsec.Maybe(maybenode, ytok_equal),
		lotprice.Yledger(db),
		ytok_closeparan)
	ylotdate := parsec.And(
		nil,
		ytok_openbrack, Ydate(db.Year()), ytok_closebrack)

	yposting := parsec.And(
		nil,
		account.Ypostaccn(db),
		parsec.Maybe(maybenode, comm.Yledger(db)),
		parsec.Maybe(maybenode, ylotprice),
		parsec.Maybe(maybenode, ylotdate),
		parsec.Maybe(maybenode, atprice.Yatprice(db)),
		parsec.Maybe(maybenode, ytok_postnote),
	)

	y := parsec.OrdChoice(
		func(nodes []parsec.ParsecNode) parsec.ParsecNode {
			switch items := nodes[0].(type) {
			case []parsec.ParsecNode:
				// account
				account := items[0].(*Account)
				accname := db.Applyroot(db.LookupAlias(account.name))
				p.account = db.GetAccount(accname).(*Account)
				p.account.virtual = account.virtual
				p.account.balanced = account.balanced

				// first commodity
				commodity, _ := items[1].(*Commodity)
				if commodity != nil { // setup commodity profiles
					p.commodity = db.GetCommodity(
						commodity.name, commodity,
					).Similar(commodity.amount)
				}

				// lot price
				lotnodes, _ := items[2].([]parsec.ParsecNode)
				if lotnodes != nil {
					equalt, ok := lotnodes[1].(*parsec.Terminal)
					if ok && equalt.Name == "EQUAL" {
						p.fixprice = lotnodes[2].(*Commodity)
					} else {
						p.lotprice = lotnodes[2].(*Commodity)
					}
				}

				// lot date
				lotnodes, _ = items[3].([]parsec.ParsecNode)
				if lotnodes != nil {
					p.lotdate = lotnodes[1].(time.Time)
				}

				// atprice
				atnodes, _ := items[4].([]parsec.ParsecNode)
				if atnodes != nil {
					at := atnodes[0].(*parsec.Terminal)
					atprice := atnodes[1].(*Commodity)
					if atprice.currency == false {
						return fmt.Errorf("at price should be currency")
					}
					if at.Name == "POSTATAT" {
						atprice.amount /= commodity.amount
					}
					p.atprice = atprice
				}

				// optionally tags or tagkv or note
				if note, ok := items[5].(*parsec.Terminal); ok {
					scanner := parsec.NewScanner([]byte(note.Value))
					if node, _ := NewTag().Yledger(db)(scanner); node == nil {
						p.note = string(note.Value)

					} else {
						tag := node.(*Tags)
						p.tags = append(p.tags, tag.tags...)
						for k, v := range tag.tagm {
							p.metadata[k] = v
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

//---- engine

func (p *Posting) TryAtPrice() *Commodity {
	if p.atprice != nil && p.commodity.currency == false {
		return p.atprice.Similar(p.commodity.amount * p.atprice.amount)
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
