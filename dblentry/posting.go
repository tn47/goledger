package dblentry

import "fmt"

import "github.com/prataprc/goparsec"
import "github.com/prataprc/golog"
import "github.com/prataprc/goledger/api"

type Posting struct {
	account   *Account
	commodity *Commodity
	note      string
}

func NewPosting() *Posting {
	return &Posting{}
}

//---- accessor

func (p *Posting) Commodity() api.Commoditiser {
	return p.commodity
}

func (p *Posting) Account() api.Accounter {
	return p.account
}

//---- ledger parser

func (p *Posting) Yledger(db *Datastore) parsec.Parser {
	account := NewAccount("")
	comm := NewCommodity("")
	yposting := parsec.And(
		nil,
		account.Ypostaccn(db),
		parsec.Maybe(maybenode, comm.Yledger(db)),
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
				// optionally note
				if note, ok := items[2].(*parsec.Terminal); ok {
					p.note = string(note.Value)
				}

				fmsg := "posting.yledger account:%v commodity:%v\n"
				log.Debugf(fmsg, p.account, p.commodity)
				return p

			case *parsec.Terminal:
				note := Transnote(string(items.Value))
				log.Debugf("posting.yledger %v\n", note)
				return note
			}
			fmsg := "unreachable code posting: len(nodes): %v"
			panic(fmt.Errorf(fmsg, len(nodes)))
		},
		yposting,
		ytok_persnote,
	)
	return y
}

//---- engine

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
