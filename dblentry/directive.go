package dblentry

import "fmt"
import "strconv"

import "github.com/prataprc/goparsec"
import "github.com/prataprc/golog"

// Directive can handle all directives in ledger journal.
type Directive struct {
	dtype      string
	year       int      // year
	account    *Account // account, alias, apply
	aliasname  string   // alias
	expression string   // assert
	endargs    []string // end
}

// NewDirective create a new Directive instance, one instance to be created
// for handling each directive in the journal file.
func NewDirective() *Directive {
	return &Directive{account: NewAccount("")}
}

//---- ledger parser

// Yledger return a parser-combinator that can parse directives.
func (d *Directive) Yledger(db *Datastore) parsec.Parser {
	y := parsec.OrdChoice(
		Vector2scalar,
		d.yaccount(db),
		d.yapply(db),
		d.yalias(db),
		d.yassert(db),
		d.yend(db),
		d.yyear(db),
	)
	return y
}

// Yledgerblock return a parser-combinator that can parse sub directives under
// account directive.
func (d *Directive) Yledgerblock(db *Datastore, block []string) error {
	var node parsec.ParsecNode
	switch d.dtype {
	case "account":
		for _, line := range block {
			scanner := parsec.NewScanner([]byte(line))
			parser := d.yaccountdirectives(db)
			if parser == nil {
				continue
			}
			node, _ = parser(scanner)
			nodes := node.([]parsec.ParsecNode)
			t := nodes[0].(*parsec.Terminal)
			switch t.Name {
			case "DRTV_ACCOUNT_NOTE":
				d.account.note = string(nodes[1].(*parsec.Terminal).Value)
			case "DRTV_ACCOUNT_ALIAS":
				d.account.aliasname = string(nodes[1].(*parsec.Terminal).Value)
				db.addAlias(d.account.aliasname, d.account.name)
			case "DRTV_ACCOUNT_PAYEE":
				d.account.payee = string(nodes[1].(*parsec.Terminal).Value)
				db.addPayee(d.account.payee, d.account.name)
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
		return nil

	case "apply", "alias", "assert", "end", "year":
		return nil
	}
	panic(fmt.Errorf("unreachable code"))
}

func (d *Directive) yaccount(db *Datastore) parsec.Parser {
	return parsec.And(
		func(nodes []parsec.ParsecNode) parsec.ParsecNode {
			d.dtype = "account"
			d.account = nodes[1].(*Account)
			log.Debugf("directive %q %v\n", d.dtype, d.account)
			return d
		},
		ytokAccount, d.account.Yledger(db),
	)
}

func (d *Directive) yapply(db *Datastore) parsec.Parser {
	return parsec.And(
		func(nodes []parsec.ParsecNode) parsec.ParsecNode {
			d.dtype = "apply"
			return d
		},
		ytokApply, ytokAccount, d.account.Yledger(db),
	)
}

func (d *Directive) yalias(db *Datastore) parsec.Parser {
	return parsec.And(
		func(nodes []parsec.ParsecNode) parsec.ParsecNode {
			d.dtype = "alias"
			d.aliasname = string(nodes[1].(*parsec.Terminal).Value)
			return d
		},
		ytokAlias, ytokAliasname, ytokEqual, d.account.Yledger(db),
	)
}

func (d *Directive) yassert(db *Datastore) parsec.Parser {
	return parsec.And(
		func(nodes []parsec.ParsecNode) parsec.ParsecNode {
			d.dtype = "assert"
			d.expression = string(nodes[1].(*parsec.Terminal).Value)
			return nil
		},
		ytokAssert, ytokExpr,
	)
}

func (d *Directive) yend(db *Datastore) parsec.Parser {
	return parsec.And(
		func(nodes []parsec.ParsecNode) parsec.ParsecNode {
			d.dtype = "end"
			d.endargs = []string{"apply", "account"}
			return d
		},
		ytokEnd, ytokApply, ytokAccount,
	)
}

func (d *Directive) yyear(db *Datastore) parsec.Parser {
	return parsec.And(
		func(nodes []parsec.ParsecNode) parsec.ParsecNode {
			d.dtype = "year"
			d.year, _ = strconv.Atoi(string(nodes[1].(*parsec.Terminal).Value))
			return d
		},
		ytokYear, ytokYearval,
	)
}

func (d *Directive) yaccountdirectives(db *Datastore) parsec.Parser {
	ynote := parsec.And(nil, ytokNote, ytokValue)
	yalias := parsec.And(nil, ytokAlias, ytokValue)
	ypayee := parsec.And(nil, ytokPayee, ytokValue)
	ycheck := parsec.And(nil, ytokCheck, ytokValue)
	yassert := parsec.And(nil, ytokAssert, ytokValue)
	yeval := parsec.And(nil, ytokEval, ytokValue)
	ydefault := parsec.And(nil, ytokDefault)
	y := parsec.OrdChoice(
		Vector2scalar,
		ynote, yalias, ypayee, ycheck, yassert, yeval, ydefault,
	)
	return y
}

//---- engine

func (d *Directive) Firstpass(db *Datastore) error {
	switch d.dtype {
	case "account":
		return db.declare(d.account) // NOTE: this is redundant

	case "apply":
		if db.rootaccount != "" {
			fmsg := "previous `apply` directive(%v) not closed"
			return fmt.Errorf(fmsg, db.rootaccount)
		}
		db.rootaccount = d.account.name
		return nil

	case "alias":
		db.addAlias(d.aliasname, d.account.name)
		return nil

	case "assert":
		return fmt.Errorf("directive not-implemented")

	case "end":
		if db.rootaccount == "" {
			return fmt.Errorf("dangling `end` directive")
		}
		db.rootaccount = ""
		return nil

	case "year":
		db.setYear(d.year)
		return nil
	}
	panic("unreachable code")
}

func (d *Directive) Secondpass(db *Datastore) error {
	return nil
}
