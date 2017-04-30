package main

import "github.com/prataprc/goparsec"

//---- Directives
var ytok_account = parsec.Token("account", "DRTV_ACCOUNT")

//---- Transaction tokens
type Transprefix byte
type Transcode string
type Transnote string

var ytok_equal = parsec.Token("=", "EQUAL")
var ytok_prefix = parsec.Maybe(
	func(nodes []parsec.ParsecNode) parsec.ParsecNode {
		s := string(nodes[0].(*parsec.Terminal).Value)
		return Transprefix(s[0])
	},
	parsec.Token(`\*|!`, "TRANSPREFIX"),
)
var ytok_code = parsec.Maybe(
	func(nodes []parsec.ParsecNode) parsec.ParsecNode {
		code := string(nodes[0].(*parsec.Terminal).Value)
		ln := len(code)
		return Transcode(code[1 : ln-1])
	},
	parsec.Token(`\(.*\)`, "TRANSCODE"),
)
var ytok_desc = parsec.Token(".+", "TRANSDESC")
var ytok_persnote = parsec.Token(";[^;]+", "TRANSPNOTE")

//---- Posting tokens

var ytok_postamount = parsec.Token("[^;]+", "AMOUNT")
var ytok_postnote = parsec.Token(";[^;]+", "TRANSNOTE")

//
func maybenode(nodes []parsec.ParsecNode) parsec.ParsecNode {
	return nodes[0]
}

func vector2scalar(nodes []parsec.ParsecNode) parsec.ParsecNode {
	return nodes[0]
}
