package dblentry

import "time"
import "fmt"
import "strings"

import "github.com/prataprc/goparsec"
import "github.com/prataprc/golog"
import "github.com/prataprc/goledger/api"

var _ = fmt.Sprintf("dummy")

type Transaction struct {
	date     time.Time
	edate    time.Time
	prefix   byte
	code     string
	desc     string
	postings []*Posting
	note     string
}

func NewTransaction() *Transaction {
	trans := &Transaction{}
	return trans
}

//---- accessor

func (trans *Transaction) Description() string {
	return trans.desc
}

func (trans *Transaction) Date() time.Time {
	return trans.date
}

func (trans *Transaction) GetPostings() []api.Poster {
	postings := []api.Poster{}
	for _, p := range trans.postings {
		postings = append(postings, p)
	}
	return postings
}

//---- ledger parser

func (trans *Transaction) Yledger(db *Datastore) parsec.Parser {
	// DATE
	ydate := Ydate(db.Year())
	// [=EDATE]
	yedate := parsec.And(
		func(nodes []parsec.ParsecNode) parsec.ParsecNode {
			return nodes[1] // EDATE
		},
		ytok_equal,
		ydate,
	)

	y := parsec.And(
		func(nodes []parsec.ParsecNode) parsec.ParsecNode {
			trans.date = nodes[0].(time.Time)
			if edate, ok := nodes[1].(time.Time); ok {
				trans.edate = edate
			}
			if t, ok := nodes[2].(*parsec.Terminal); ok {
				trans.prefix = t.Value[0]
			}
			if t, ok := nodes[3].(*parsec.Terminal); ok {
				trans.code = string(t.Value[1 : len(t.Value)-1])
			}
			trans.desc = string(nodes[4].(*parsec.Terminal).Value)
			log.Debugf("trans.yledger %v %v\n", trans.date, trans.desc)
			return trans
		},
		ydate,
		parsec.Maybe(maybenode, yedate),
		parsec.Maybe(maybenode, ytok_prefix),
		parsec.Maybe(maybenode, ytok_code),
		ytok_desc,
	)
	return y
}

func (trans *Transaction) Yledgerblock(db *Datastore, block []string) {
	var node parsec.ParsecNode

	for _, line := range block {
		scanner := parsec.NewScanner([]byte(line))
		posting := NewPosting()
		node, scanner = posting.Yledger(db)(scanner)
		switch val := node.(type) {
		case *Posting:
			trans.postings = append(trans.postings, val)
		case Transnote:
			trans.note = string(val)
		}
	}
}

//---- engine

func (trans *Transaction) ShouldBalance() bool {
	for _, posting := range trans.postings {
		if posting.virtual == true && posting.balanced == false {
			return false
		} else if posting.balanced == false {
			return false
		}
	}
	return true
}

func (trans *Transaction) Defaultposting(
	db *Datastore, defacc *Account, commodity *Commodity) *Posting {

	posting := NewPosting()
	posting.account = defacc
	posting.commodity = commodity.Similar(commodity.amount)
	return posting
}

func (trans *Transaction) Endposting(postings []*Posting) (*Posting, error) {
	var tallywith *Posting
	for _, posting := range postings {
		if posting.commodity == nil && tallywith != nil {
			err := fmt.Errorf("Only one null posting allowed per transaction")
			return nil, err
		} else if posting.commodity == nil {
			tallywith = posting
		}
	}
	return tallywith, nil
}

func (trans *Transaction) Autobalance1(
	db *Datastore, defaccount *Account) (bool, error) {

	if len(trans.postings) == 0 {
		return false, fmt.Errorf("empty transaction")

	} else if len(trans.postings) == 1 && defaccount != nil {
		commodity := trans.postings[0].commodity
		posting := trans.Defaultposting(db, defaccount, commodity)
		posting.commodity.InverseAmount()
		trans.postings = append(trans.postings, posting)
		return true, nil

	} else if len(trans.postings) == 1 {
		return false, fmt.Errorf("unbalanced transaction")
	}

	tallypost, err := trans.Endposting(trans.postings)
	if err != nil {
		return false, err
	}

	unbcs := trans.DoBalance()
	if len(unbcs) == 0 {
		return true, nil
	}
	tallypost.commodity = unbcs[0]
	tallypost.commodity.InverseAmount()
	if len(unbcs) > 1 {
		account := tallypost.account
		for _, unbc := range unbcs[1:] {
			posting := trans.Defaultposting(db, account, unbc)
			posting.commodity.InverseAmount()
			trans.postings = append(trans.postings, posting)
		}
	}
	return true, nil
}

func (trans *Transaction) DoBalance() []*Commodity {
	unbalanced := map[string]*Commodity{}
	for _, posting := range trans.postings {
		if posting.commodity == nil {
			continue
		}
		unbc, ok := unbalanced[posting.commodity.name]
		if ok == false {
			unbc = posting.commodity.Similar(0)
		}
		unbc.Add(posting.commodity)
		unbalanced[unbc.name] = unbc
	}
	unbcs := []*Commodity{}
	for _, unbc := range unbalanced {
		if unbc.amount != 0 {
			unbcs = append(unbcs, unbc)
		}
	}
	return unbcs
}

func (trans *Transaction) Firstpass(db *Datastore) error {
	if trans.ShouldBalance() {
		defaccount := db.GetAccount(db.blncingaccnt)
		if ok, err := trans.Autobalance1(db, defaccount); err != nil {
			return err
		} else if ok == false {
			return fmt.Errorf("unbalanced transaction")
		}
		log.Debugf("transaction balanced\n")
	}
	for _, posting := range trans.postings {
		if err := posting.Firstpass(db, trans); err != nil {
			return err
		}
	}
	return nil
}

func (trans *Transaction) Secondpass(db *Datastore) error {
	for _, posting := range trans.postings {
		if err := posting.Secondpass(db, trans); err != nil {
			return err
		}
	}
	return db.reporter.Transaction(db, trans)
}

func FitDescription(desc string, maxwidth int) string {
	if len(desc) < maxwidth {
		return desc
	}
	scraplen := len(desc) - maxwidth
	fields := []string{}
	for _, field := range strings.Fields(desc) {
		if scraplen <= 0 || len(field) <= 3 {
			fields = append(fields, field)
			continue
		}
		if len(field[3:]) < scraplen {
			fields = append(fields, field[:3])
			scraplen -= len(field[3:])
			continue
		}
		fields = append(fields, field[:len(field)-scraplen])
		scraplen = 0
	}
	return strings.Join(fields, " ")
}
