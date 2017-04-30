package main

import "github.com/prataprc/goparsec"

type Directive struct {
	dtype       string
	accountname string   // account, alias
	account     *Account // account
	applyname   string   // apply
	aliasname   string   // alias
}

func NewDirective() *Directive {
	return &Directive{}
}

func (d *Directive) Y(db *Datastore) parsec.Parser {
	y := parsec.OrdChoice(
		vector2scalar,
		d.Yaccount(db),
		d.Yapply(db),
	)
	return y
}

func (d *Directive) Yaccount(db *Datastore) parsec.Parser {
	d.account = NewAccount("")
	return parsec.And(
		func(nodes []parsec.ParsecNode) parsec.ParsecNode {
			d.accountname = nodes[1].(*Account).Name()
			account := db.GetAccount(d.accountname, true /*declare*/)
			if account != nil {
				d.account = account
			}
			d.dtype = "account"
			return d
		},
		ytok_account, d.account.Y(db),
	)
}

func (d *Directive) Yapply(db *Datastore) parsec.Parser {
	d.account = NewAccount("")
	return parsec.And(
		func(nodes []parsec.ParsecNode) parsec.ParsecNode {
			d.applyname = nodes[2].(*Account).Name()
			d.dtype = "apply"
			return d
		},
		ytok_apply, ytok_account, d.account.Y(db),
	)
}

func (d *Directive) Yalias(db *Datastore) parsec.Parser {
	d.account = NewAccount("")
	return parsec.And(
		func(nodes []parsec.ParsecNode) parsec.ParsecNode {
			d.aliasname = string(nodes[1].(*parsec.Terminal).Value)
			d.accountname = nodes[3].(*Account).Name()
			d.dtype = "apply"
			return d
		},
		ytok_alias, ytok_aliasname, ytok_equal, d.account.Y(db),
	)
}

func (d *Directive) Yattr(db *Datastore) parsec.Parser {
	switch d.dtype {
	case "account":
		ynote := parsec.And(nil, ytok_note, ytok_value)
		yalias := parsec.And(nil, ytok_alias, ytok_value)
		ypayee := parsec.And(nil, ytok_payee, ytok_value)
		ycheck := parsec.And(nil, ytok_check, ytok_value)
		yassert := parsec.And(nil, ytok_assert, ytok_value)
		yeval := parsec.And(nil, ytok_eval, ytok_value)
		ydefault := parsec.And(nil, ytok_default)
		y := parsec.OrdChoice(
			nil,
			ynote, yalias, ypayee, ycheck, yassert, yeval, ydefault,
		)
		return y

	case "apply", "alias":
		return nil
	}
	panic("unreachable code")
}

func (d *Directive) Apply(db *Datastore, node parsec.ParsecNode) {
	switch d.dtype {
	case "account":
		nodes := node.([]parsec.ParsecNode)
		t := nodes[0].(*parsec.Terminal)
		switch t.Name {
		case "DRTV_ACCOUNT_NOTE":
			d.account.SetNote(string(nodes[1].(*parsec.Terminal).Value))
		case "DRTV_ACCOUNT_ALIAS":
			aliasname := string(nodes[1].(*parsec.Terminal).Value)
			db.AddAlias(aliasname, d.accountname)
		case "DRTV_ACCOUNT_PAYEE":
			d.account.SetPayee(string(nodes[1].(*parsec.Terminal).Value))
		case "DRTV_ACCOUNT_CHECK":
			d.account.SetCheck(string(nodes[1].(*parsec.Terminal).Value))
		case "DRTV_ACCOUNT_ASSERT":
			d.account.SetAssert(string(nodes[1].(*parsec.Terminal).Value))
		case "DRTV_ACCOUNT_EVAL":
			d.account.SetEval(string(nodes[1].(*parsec.Terminal).Value))
		case "DRTV_ACCOUNT_DEFAULT":
			db.SetBalancingaccount(d.account)
		}

	case "apply", "alias":
		return
	}
	panic("unreachable code")
}
