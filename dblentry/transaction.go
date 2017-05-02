package dblentry

import "time"
import "fmt"

import "github.com/prataprc/goparsec"
import "github.com/prataprc/golog"

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

//---- Legder text combinators

func (trans *Transaction) Yledger(db *Datastore) parsec.Parser {
	// DATE
	ydate := Ydate(db.Year(), db.Month(), db.Dateformat())
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
