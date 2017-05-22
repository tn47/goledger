package dblentry

import "fmt"
import "strconv"

import "github.com/prataprc/goparsec"
import "github.com/prataprc/golog"

// Directive can handle all directives in ledger journal.
type Directive struct {
	dtype      string
	year       int    // year
	note       string // account, commodity
	ndefault   bool   // account, commodity
	accname    string // account, alias, apply
	accalias   string // account
	accpayee   string // account
	acccheck   string // account
	accassert  string // account
	acceval    string // account
	aliasname  string // alias
	expression string // assert, check
	capture    string // capture pattern
	commdname  string // commodity
	commdfmt   string
	commdnmrkt bool
	endargs    []string // end
	comments   []string
}

// NewDirective create a new Directive instance, one instance to be created
// for handling each directive in the journal file.
func NewDirective() *Directive {
	return &Directive{}
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
		d.ybucket(db),
		d.ycheck(db),
		d.ycapture(db),
		d.ycheck(db),
		d.ycomment(db),
		d.ycommodity(db),
		d.yend(db),
		d.yyear(db),
	)
	return y
}

// Yledgerblock return a parser-combinator that can parse sub directives under
// account directive.
func (d *Directive) Yledgerblock(db *Datastore, block []string) (int, error) {
	var node parsec.ParsecNode
	switch d.dtype {
	case "account":
		for index, line := range block {
			scanner := parsec.NewScanner([]byte(line))
			parser := d.yaccountdirectives(db)
			if parser == nil {
				continue
			}
			node, _ = parser(scanner)
			if node == nil {
				return index, fmt.Errorf("parsing %q", line)
			}
			nodes := node.([]parsec.ParsecNode)
			t := nodes[0].(*parsec.Terminal)
			switch t.Name {
			case "DRTV_NOTE":
				d.note = string(nodes[2].(*parsec.Terminal).Value)
			case "DRTV_ACCOUNT_ALIAS":
				d.accalias = string(nodes[2].(*parsec.Terminal).Value)
			case "DRTV_ACCOUNT_PAYEE":
				d.accpayee = string(nodes[2].(*parsec.Terminal).Value)
			case "DRTV_ACCOUNT_CHECK":
				d.acccheck = string(nodes[2].(*parsec.Terminal).Value)
			case "DRTV_ACCOUNT_ASSERT":
				d.accassert = string(nodes[2].(*parsec.Terminal).Value)
			case "DRTV_ACCOUNT_EVAL":
				d.acceval = string(nodes[2].(*parsec.Terminal).Value)
			case "DRTV_DEFAULT":
				d.ndefault = true
			}
		}
		return len(block), nil

	case "commodity":
		for index, line := range block {
			scanner := parsec.NewScanner([]byte(line))
			parser := d.ycommoditydirectives(db)
			if parser == nil {
				continue
			}
			node, _ = parser(scanner)
			if node == nil {
				return index, fmt.Errorf("parsing %q", line)
			}
			nodes := node.([]parsec.ParsecNode)
			t := nodes[0].(*parsec.Terminal)
			switch t.Name {
			case "DRTV_NOTE":
				d.note = nodes[2].(*parsec.Terminal).Value
			case "DRTV_COMMODITY_FORMAT":
				d.commdfmt = nodes[2].(*parsec.Terminal).Value
			case "DRTV_COMMODITY_NOMARKET":
				d.commdnmrkt = true
			case "DRTV_DEFAULT":
				d.ndefault = true
			}
		}
		return len(block), nil

	case "apply", "alias", "assert", "bucket", "capture", "check", "comment",
		"end", "year":
		return len(block), nil
	}
	panic(fmt.Errorf("unreachable code"))
}

func (d *Directive) yaccount(db *Datastore) parsec.Parser {
	account := NewAccount("") // local scope.
	return parsec.And(
		func(nodes []parsec.ParsecNode) parsec.ParsecNode {
			d.dtype = "account"
			d.accname = nodes[1].(*Account).name
			log.Debugf("directive %q %v\n", d.dtype, d.accname)
			return d
		},
		ytokAccount, account.Yledger(db),
	)
}

func (d *Directive) yapply(db *Datastore) parsec.Parser {
	account := NewAccount("") // local scope.
	return parsec.And(
		func(nodes []parsec.ParsecNode) parsec.ParsecNode {
			d.dtype = "apply"
			d.accname = nodes[2].(*Account).name
			return d
		},
		ytokApply, ytokAccount, account.Yledger(db),
	)
}

func (d *Directive) yalias(db *Datastore) parsec.Parser {
	account := NewAccount("") // local scope.
	return parsec.And(
		func(nodes []parsec.ParsecNode) parsec.ParsecNode {
			d.dtype = "alias"
			d.aliasname = string(nodes[1].(*parsec.Terminal).Value)
			d.accname = nodes[3].(*Account).name
			return d
		},
		ytokAlias, ytokAliasname, ytokEqual, account.Yledger(db),
	)
}

func (d *Directive) yassert(db *Datastore) parsec.Parser {
	return parsec.And(
		func(nodes []parsec.ParsecNode) parsec.ParsecNode {
			d.dtype = "assert"
			d.expression = string(nodes[1].(*parsec.Terminal).Value)
			return d
		},
		ytokAssert, ytokExpr,
	)
}

func (d *Directive) ybucket(db *Datastore) parsec.Parser {
	account := NewAccount("") // local scope
	return parsec.And(
		func(nodes []parsec.ParsecNode) parsec.ParsecNode {
			d.dtype = "bucket"
			d.accname = nodes[1].(*Account).name
			return d
		},
		ytokBucket, account.Yledger(db),
	)
}

func (d *Directive) ycapture(db *Datastore) parsec.Parser {
	account := NewAccount("") // local scope
	return parsec.And(
		func(nodes []parsec.ParsecNode) parsec.ParsecNode {
			d.dtype = "capture"
			d.accname = nodes[1].(*Account).name
			d.capture = nodes[2].(*parsec.Terminal).Value
			return d
		},
		ytokCapture, account.Ypostaccn(db), ytokValue,
	)
}

func (d *Directive) ycheck(db *Datastore) parsec.Parser {
	return parsec.And(
		func(nodes []parsec.ParsecNode) parsec.ParsecNode {
			d.dtype = "check"
			d.expression = nodes[1].(*parsec.Terminal).Value
			return d
		},
		ytokCheck, ytokExpr,
	)
}

func (d *Directive) ycomment(db *Datastore) parsec.Parser {
	return parsec.And(
		func(nodes []parsec.ParsecNode) parsec.ParsecNode {
			d.dtype = "comment"
			return d
		},
		ytokComment,
	)
}

func (d *Directive) ycommodity(db *Datastore) parsec.Parser {
	commodity := NewCommodity("") // local scope.
	return parsec.And(
		func(nodes []parsec.ParsecNode) parsec.ParsecNode {
			d.dtype = "commodity"
			d.commdname = nodes[1].(*parsec.Terminal).Value
			log.Debugf("directive %q %v\n", d.dtype, d.commdname)
			return d
		},
		ytokDirtCommodity, commodity.Yname(db),
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
	ynote := parsec.And(nil, ytokNote, ytokHardSpace, ytokValue)
	yalias := parsec.And(nil, ytokAlias, ytokHardSpace, ytokValue)
	ypayee := parsec.And(nil, ytokPayee, ytokHardSpace, ytokValue)
	ycheck := parsec.And(nil, ytokCheck, ytokHardSpace, ytokValue)
	yassert := parsec.And(nil, ytokAssert, ytokHardSpace, ytokValue)
	yeval := parsec.And(nil, ytokEval, ytokHardSpace, ytokValue)
	ydefault := parsec.And(nil, ytokDefault)
	y := parsec.OrdChoice(
		Vector2scalar,
		ynote, yalias, ypayee, ycheck, yassert, yeval, ydefault,
	)
	return y
}

func (d *Directive) ycommoditydirectives(db *Datastore) parsec.Parser {
	ynote := parsec.And(nil, ytokNote, ytokHardSpace, ytokValue)
	yformat := parsec.And(nil, ytokFormat, ytokHardSpace, ytokValue)
	ynomarket := parsec.And(nil, ytokNomarket)
	ydefault := parsec.And(nil, ytokDefault)
	y := parsec.OrdChoice(Vector2scalar, ynote, yformat, ynomarket, ydefault)
	return y
}

//---- engine

func (d *Directive) Firstpass(db *Datastore) error {
	switch d.dtype {
	case "account":
		return db.declare(d)

	case "apply":
		return db.setrootaccount(d.accname)

	case "alias":
		if db.IsStrict() && db.HasAccount(d.accname) == false {
			return fmt.Errorf("account %q not declared before", d.accname)
		}
		db.addAlias(d.aliasname, d.accname)
		return nil

	case "assert":
		return fmt.Errorf("assert directive not-implemented")

	case "bucket":
		if db.IsStrict() && db.HasAccount(d.accname) == false {
			return fmt.Errorf("account %q not declared before", d.accname)
		}
		db.setBalancingaccount(d.accname)
		return nil

	case "capture":
		if db.IsStrict() && db.HasAccount(d.accname) == false {
			return fmt.Errorf("account %q not declared before", d.accname)
		}
		db.addCapture(d.capture, d.accname)
		return nil

	case "check":
		return fmt.Errorf("check directive not-implemented")

	case "comment":
		return fmt.Errorf("comment directive not-implemented")

	case "commodity":
		return db.declare(d)

	case "end":
		return db.clearRootaccount()

	case "year":
		db.setYear(d.year)
		return nil
	}
	panic("unreachable code")
}

func (d *Directive) Secondpass(db *Datastore) error {
	return nil
}
