package dblentry

import "fmt"

import "github.com/prataprc/goparsec"

type Directive struct {
	dtype      string
	year       int      // year
	month      int      // month
	dateformat string   // dateformat
	account    *Account // account, alias, apply
	aliasname  string   // alias
	expression string   // assert
}

func NewDirective() *Directive {
	return &Directive{account: NewAccount("")}
}

//---- ledger parser

func (d *Directive) Yledger(db *Datastore) parsec.Parser {
	y := parsec.OrdChoice(
		Vector2scalar,
		d.Yaccount(db),
		d.Yapply(db),
		d.Yalias(db),
		d.Yassert(db),
	)
	return y
}

func (d *Directive) Yaccount(db *Datastore) parsec.Parser {
	return parsec.And(
		func(nodes []parsec.ParsecNode) parsec.ParsecNode {
			d.dtype = "account"
			d.account = nodes[1].(*Account)
			return d
		},
		ytok_account, d.account.Yledger(db),
	)
}

func (d *Directive) Yapply(db *Datastore) parsec.Parser {
	return parsec.And(
		func(nodes []parsec.ParsecNode) parsec.ParsecNode {
			d.dtype = "apply"
			return d
		},
		ytok_apply, ytok_account, d.account.Yledger(db),
	)
}

func (d *Directive) Yalias(db *Datastore) parsec.Parser {
	return parsec.And(
		func(nodes []parsec.ParsecNode) parsec.ParsecNode {
			d.dtype = "apply"
			d.aliasname = string(nodes[1].(*parsec.Terminal).Value)
			return d
		},
		ytok_alias, ytok_aliasname, ytok_equal, d.account.Yledger(db),
	)
}

func (d *Directive) Yassert(db *Datastore) parsec.Parser {
	return parsec.And(
		func(nodes []parsec.ParsecNode) parsec.ParsecNode {
			d.dtype = "assert"
			d.expression = string(nodes[1].(*parsec.Terminal).Value)
			return nil
		},
		ytok_assert, ytok_expr,
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
			Vector2scalar,
			ynote, yalias, ypayee, ycheck, yassert, yeval, ydefault,
		)
		return y

	case "apply", "alias", "assert":
		return nil
	}
	panic("unreachable code")
}

func (d *Directive) Yledgerblock(db *Datastore, block []string) {
	var node parsec.ParsecNode
	switch d.dtype {
	case "account":
		for _, line := range block {
			scanner := parsec.NewScanner([]byte(line))
			parser := d.Yattr(db)
			if parser == nil {
				continue
			}
			node, scanner = parser(scanner)
			nodes := node.([]parsec.ParsecNode)
			t := nodes[0].(*parsec.Terminal)
			switch t.Name {
			case "DRTV_ACCOUNT_NOTE":
				d.account.note = string(nodes[1].(*parsec.Terminal).Value)
			case "DRTV_ACCOUNT_ALIAS":
				aliasname := string(nodes[1].(*parsec.Terminal).Value)
				db.AddAlias(aliasname, d.account.name)
			case "DRTV_ACCOUNT_PAYEE":
				payee := string(nodes[1].(*parsec.Terminal).Value)
				db.AddPayee(payee, d.account.name)
			case "DRTV_ACCOUNT_CHECK":
				d.account.check = string(nodes[1].(*parsec.Terminal).Value)
			case "DRTV_ACCOUNT_ASSERT":
				d.account.assert = string(nodes[1].(*parsec.Terminal).Value)
			case "DRTV_ACCOUNT_EVAL":
				d.account.eval = string(nodes[1].(*parsec.Terminal).Value)
			case "DRTV_ACCOUNT_DEFAULT":
				d.account.defblns = true
			}
		}
		return

	case "apply", "alias":
		return
	}
	panic(fmt.Errorf("unreachable code"))
}
