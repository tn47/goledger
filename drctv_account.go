package main

import "github.com/prataprc/goparsec"

type DirectiveAccount struct {
	name    string
	note    string
	descrip string
	alias   string
	payee   string
	check   string
	assert  string
	eval    string

	context Context
}

func NewDirectiveAccount(context Context) *DirectiveAccount {
	return &DirectiveAccount{context: context}
}

func (drtv *DirectiveAccount) Y() parsec.Parser {
	account := NewAccount("", drtv.context)
	// account
	ybegin := parsec.Token("account", "DRTV_ACCOUNT")
	y := parsec.And(
		func(nodes []parsec.ParsecNode) parsec.ParsecNode {
			return nil
		},
		ybegin,
		account.Y(),
	)
	return y
}

func (drtv *DirectiveAccount) Parsedirective(scanner parsec.Scanner) {
}
