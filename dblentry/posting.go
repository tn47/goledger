package dblentry

import "fmt"
import "strings"

import "github.com/prataprc/goparsec"
import "github.com/prataprc/golog"
import "github.com/prataprc/goledger/api"

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
		account.Yledger(db),
		parsec.Maybe(maybenode, comm.Yledger(db)),
		parsec.Maybe(maybenode, ytok_postnote),
	)

	// hack to fix an issue with regexp issue #1
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
			var prefix string

			node := nodes[0]
			switch items := node.(type) {
			case []parsec.ParsecNode:
				// first commodity
				commodity, _ := items[1].(*Commodity)
				// back to account name
				account := items[0].(*Account)
				p.virtual = account.Virtual()
				p.balanced = account.Balanced()
				accname := db.Applyroot(db.LookupAlias(account.name))
				// hack to fix issue with regexp issue #1
				if commodity != nil && commodity.name == "" {
					accname, prefix = getamountprefix(accname)
					if prefix == "-" {
						commodity.amount = -commodity.amount
					} else {
						scanner := parsec.NewScanner([]byte(prefix))
						node, _ := ytok_currency(scanner)
						if t, ok := node.(*parsec.Terminal); ok {
							commodity.name = string(t.Value)
							commodity.currency = true
						}
					}
				}
				// setup commodity profiles
				if commodity != nil {
					p.commodity = db.GetCommodity(
						commodity.name, commodity).Similar(commodity.amount)
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
	if err := p.account.Secondpass(db, trans, p); err != nil {
		return err
	}
	if err := p.commodity.Secondpass(db, trans, p); err != nil {
		return err
	}
	return nil
}
