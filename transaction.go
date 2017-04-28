package main

import "time"

import "github.com/prataprc/goparsec"
import s "github.com/prataprc/gosettings"

type Transprefix byte
type Transcode string

type Transaction struct {
	// start
	date     time.Time
	edate    time.Time
	prefix   byte
	code     string
	desc     string
	postings []*Posting
	note     string

	// context
	year       int
	month      int
	dateformat string
	context    s.Settings
}

func NewTransaction(context s.Settings) *Transaction {
	trans := &Transaction{
		year:       int(context.Int64("year")),
		month:      int(context.Int64("month")),
		dateformat: context.String("dateformat"),
		context:    context,
	}
	return trans
}

func (trans *Transaction) Y() parsec.Parser {
	edater := func(nodes []parsec.ParsecNode) parsec.ParsecNode {
		return nodes[1] // EDATE
	}

	prefixer := func(nodes []parsec.ParsecNode) parsec.ParsecNode {
		s := string(nodes[0].(*parsec.Terminal).Value)
		return Transprefix(s[0])
	}

	coder := func(nodes []parsec.ParsecNode) parsec.ParsecNode {
		code := string(nodes[0].(*parsec.Terminal).Value)
		ln := len(code)
		return Transcode(code[1 : ln-1])
	}

	noder := func(nodes []parsec.ParsecNode) parsec.ParsecNode {
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
		return nil
	}

	// DATE
	ydate := Ydate(trans.year, trans.month, trans.dateformat)
	// [=EDATE]
	yequal := parsec.Token("=", "TRANSEQUAL")
	yedate := parsec.Maybe(maybenode, parsec.And(edater, yequal, ydate))
	// [*|!]
	yprefix := parsec.Maybe(prefixer, parsec.Token(`\*|!`, "TRANSPREFIX"))
	// [(CODE)]
	ycode := parsec.Maybe(coder, parsec.Token(`\(.*\)`, "TRANSCODE"))
	// DESC
	ydesc := parsec.Token(".+", "TRANSDESC")

	y := parsec.And(noder, ydate, yedate, yprefix, ycode, ydesc)
	return y
}

func maybenode(nodes []parsec.ParsecNode) parsec.ParsecNode {
	return nodes[0]
}
