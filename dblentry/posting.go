package dblentry

import "fmt"
import "strings"

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
	return &Posting{virtual: false, balanced: true}
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

	// hack to fix an issue with regexp
	getamountprefix := func(name string) (string, string) {
		lastbyte := name[len(name)-1]
		if lastbyte == ' ' || lastbyte == '\t' {
			return name, ""
		}
		j := len(name) - 1
		for ; j >= 0; j-- {
			if name[j] == ' ' || name[j] == '\t' {
				break
			}
		}
		return name[:j+1], name[j+1:]
	}

	y := parsec.OrdChoice(
		func(nodes []parsec.ParsecNode) parsec.ParsecNode {
			node := nodes[0]
			switch items := node.(type) {
			case []parsec.ParsecNode:
				// first commodity
				if commodity, ok := items[1].(*Commodity); ok {
					p.commodity = commodity
				}
				// back to account name
				account := items[0].(*Account)
				p.virtual = account.Virtual()
				p.balanced = account.Balanced()
				accname := db.Applyroot(db.LookupAlias(account.name))
				if p.commodity != nil && p.commodity.name == "" {
					accname, p.commodity.name = getamountprefix(accname)
				}
				accname = strings.TrimRight(accname, " \t")
				p.account = db.GetAccount(accname)
				// optionally note
				if note, ok := items[2].(*parsec.Terminal); ok {
					p.note = string(note.Value)
				}

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

func (p *Posting) Apply(db *Datastore, trans *Transaction) {
	p.account.Apply(db, trans, p)
}
