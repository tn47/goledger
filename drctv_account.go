package main

import "github.com/prataprc/goparsec"

type DirectiveAccount struct {
	account *Account
	db      *Datastore // read-only copy
}

func NewDirectiveAccount(db *Datastore) *DirectiveAccount {
	return &DirectiveAccount{db: db}
}

func (drtv *DirectiveAccount) Y() parsec.Parser {
	drtv.account = NewAccount("", drtv.db)
	y := parsec.And(
		func(nodes []parsec.ParsecNode) parsec.ParsecNode {
			account := drtv.db.GetAccount(nodes[1].(*Account).Name())
			if account != nil {
				drtv.account = account
			}
			return drtv
		},
		parsec.Token("account", "DRTV_ACCOUNT"), // account
		drtv.account.Y(),
	)
	return y
}

func (drtv *DirectiveAccount) Parse(scanner parsec.Scanner) parsec.Scanner {
	var bs []byte
	var pn parsec.ParsecNode

	account := drtv.account

	for {
		if bs, scanner = scanner.SkipWS(); len(bs) == 0 {
			return scanner
		}
		pn, scanner = drtv.subdirective()(scanner)
		nodes := pn.([]parsec.ParsecNode)
		t := nodes[0].(*parsec.Terminal)
		switch t.Name {
		case "DRTV_ACCOUNT_NOTE":
			account.SetNote(string(nodes[1].(*parsec.Terminal).Value))
		case "DRTV_ACCOUNT_ALIAS":
			account.SetAlias(string(nodes[1].(*parsec.Terminal).Value))
		case "DRTV_ACCOUNT_PAYEE":
			account.SetPayee(string(nodes[1].(*parsec.Terminal).Value))
		case "DRTV_ACCOUNT_CHECK":
			account.SetCheck(string(nodes[1].(*parsec.Terminal).Value))
		case "DRTV_ACCOUNT_ASSERT":
			account.SetAssert(string(nodes[1].(*parsec.Terminal).Value))
		case "DRTV_ACCOUNT_EVAL":
			account.SetEval(string(nodes[1].(*parsec.Terminal).Value))
		case "DRTV_ACCOUNT_DEFAULT":
			drtv.db.SetBalancingaccount(account)
		}
	}
}

func (drtv *DirectiveAccount) subdirective() parsec.Parser {
	ynote := parsec.And(nil,
		parsec.Token("note", "DRTV_ACCOUNT_NOTE"),
		parsec.Token(".*", "DRTV_ACCOUNT_NOTE_VALUE"),
	)
	yalias := parsec.And(nil,
		parsec.Token("alias", "DRTV_ACCOUNT_ALIAS"),
		parsec.Token(".*", "DRTV_ACCOUNT_ALIAS_VALUE"),
	)
	ypayee := parsec.And(nil,
		parsec.Token("payee", "DRTV_ACCOUNT_PAYEE"),
		parsec.Token(".*", "DRTV_ACCOUNT_PAYEE_VALUE"),
	)
	ycheck := parsec.And(nil,
		parsec.Token("check", "DRTV_ACCOUNT_CHECK"),
		parsec.Token(".*", "DRTV_ACCOUNT_CHECK_VALUE"),
	)
	yassert := parsec.And(nil,
		parsec.Token("assert", "DRTV_ACCOUNT_ASSERT"),
		parsec.Token(".*", "DRTV_ACCOUNT_ASSERT_VALUE"),
	)
	yeval := parsec.And(nil,
		parsec.Token("eval", "DRTV_ACCOUNT_EVAL"),
		parsec.Token(".*", "DRTV_ACCOUNT_EVAL_VALUE"),
	)
	ydefault := parsec.And(nil,
		parsec.Token("default", "DRTV_ACCOUNT_DEFAULT"),
	)
	y := parsec.OrdChoice(
		nil,
		ynote, yalias, ypayee, ycheck, yassert, yeval, ydefault,
	)
	return y
}
