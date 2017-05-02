package dblentry

import "fmt"

import "github.com/prataprc/goparsec"
import "github.com/prataprc/golog"

type Posting struct {
	direction string // "source", target"
	// post entry
	account   *Account
	virtual   bool
	balanced  bool
	commodity *Commodity
	note      string
}

func NewPosting() *Posting {
	return &Posting{}
}

func (p *Posting) Yledger(db *Datastore) parsec.Parser {
	account := NewAccount("")
	commodity := NewCommodity()
	yposting := parsec.And(
		nil,
		account.Yledger(db),
		parsec.Maybe(maybenode, commodity.Yledger(db)),
		parsec.Maybe(maybenode, ytok_postnote),
	)

	y := parsec.OrdChoice(
		func(nodes []parsec.ParsecNode) parsec.ParsecNode {
			node := nodes[0]
			switch items := node.(type) {
			case []parsec.ParsecNode:
				p.account = items[0].(*Account)
				if commodity, ok := items[1].(*Commodity); ok {
					p.commodity = commodity
				}
				if note, ok := items[2].(*parsec.Terminal); ok {
					p.note = string(note.Value)
				}
				p.virtual = p.account.Virtual()
				p.balanced = p.account.Balanced()
				fmsg := "posting.yledger %v %v\n"
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
