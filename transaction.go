package main

import "time"

import "github.com/prataprc/goparsec"

type Transaction struct {
	// start
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

func (trans *Transaction) Y(db *Datastore) parsec.Parser {
	// DATE
	ydate := Ydate(db.Year(), db.Month(), db.Dateformat())
	// [=EDATE]
	yedate := parsec.Maybe(
		maybenode,
		parsec.And(
			func(nodes []parsec.ParsecNode) parsec.ParsecNode {
				return nodes[1] // EDATE
			},
			ytok_equal,
			ydate,
		),
	)

	y := parsec.And(
		func(nodes []parsec.ParsecNode) parsec.ParsecNode {
			n := 0
			trans.date = nodes[n].(time.Time)
			n++
			if edate, ok := nodes[n].(time.Time); ok {
				trans.edate = edate
				n++
			}
			if prefix, ok := nodes[n].(Transprefix); ok {
				trans.prefix = byte(prefix)
				n++
			}
			if code, ok := nodes[n].(Transcode); ok {
				trans.code = string(code)
				n++
			}
			trans.desc = nodes[n].(string)
			return trans
		},
		ydate, yedate, ytok_prefix, ytok_code, ytok_desc,
	)
	return y
}

func (trans *Transaction) Apply(db *Datastore, node parsec.ParsecNode) {
	switch val := node.(type) {
	case *Posting:
		trans.postings = append(trans.postings, val)
	case Transnote:
		trans.note = string(val)
	}
}
