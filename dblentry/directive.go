package dblentry

import "fmt"
import "strings"
import "strconv"

import "github.com/prataprc/goparsec"
import "github.com/prataprc/golog"
import "github.com/tn47/goledger/api"

// Directive can handle all directives in ledger journal.
type Directive struct {
	dtype       string
	year        int      // year
	note        string   // account, commodity
	comments    []string // account, commodity
	ndefault    bool     // account, commodity
	accname     string   // account, alias, apply
	accalias    string   // account
	accpayee    string   // account
	acccheck    string   // account
	accassert   string   // account
	acceval     string   // account
	acctypes    []string // account
	aliasname   string   // alias
	expression  string   // assert, check
	capture     string   // capture pattern
	commdname   string   // commodity
	commdfmt    string   // commodity
	commdnmrkt  bool     // commodity
	commdcurrn  bool     // commodity
	includefile string   // include
	dpayee      string   // payee
	dpayeealias []string // payee
	dpayeeuuid  []string // payee
	endargs     []string // end
}

// NewDirective create a new Directive instance, one instance to be created
// for handling each directive in the journal file.
func NewDirective() *Directive {
	return &Directive{}
}

//---- exported methods

func (d *Directive) Type() string {
	return d.dtype
}

func (d *Directive) Includefile() string {
	if d.dtype == "include" {
		return d.includefile
	}
	panic("impossible situation")
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
		d.ydefine(db),
		d.yfixed(db),
		d.yinclude(db),
		d.ypayee(db),
		d.ytest(db),
		d.yend(db),
		d.yyear(db),
	)
	return y
}

// Yledgerblock return a parser-combinator that can parse sub directives under
// account directive.
func (d *Directive) Yledgerblock(db *Datastore, block []string) (int, error) {
	trimstr := func(n parsec.ParsecNode) string {
		return strings.Trim(n.(*parsec.Terminal).Value, " \t")
	}

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
				d.note = trimstr(nodes[2])
			case "DRTV_SHORTNOTE":
				d.addBlockComment(trimstr(nodes[0]))
			case "DRTV_ACCOUNT_ALIAS":
				d.accalias = trimstr(nodes[2])
			case "DRTV_ACCOUNT_PAYEE":
				d.accpayee = trimstr(nodes[2])
			case "DRTV_ACCOUNT_CHECK":
				d.acccheck = trimstr(nodes[2])
			case "DRTV_ACCOUNT_ASSERT":
				d.accassert = trimstr(nodes[2])
			case "DRTV_ACCOUNT_EVAL":
				d.acceval = trimstr(nodes[2])
			case "DRTV_ACCOUNT_TYPE":
				acctypes := api.Parsecsv(trimstr(nodes[2]))
				d.acctypes = d.addAccounttype(acctypes, d.acctypes)
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
			case "DRTV_SHORTNOTE":
				d.addBlockComment(trimstr(nodes[0]))
			case "DRTV_COMMODITY_FORMAT":
				d.commdfmt = nodes[2].(*parsec.Terminal).Value
			case "DRTV_COMMODITY_NOMARKET":
				d.commdnmrkt = true
			case "DRTV_COMMODITY_CURRENCY":
				d.commdcurrn = true
			case "DRTV_DEFAULT":
				d.ndefault = true
			}
		}
		return len(block), nil

	case "payee":
		d.dpayeealias = []string{}
		d.dpayeeuuid = []string{}
		for index, line := range block {
			scanner := parsec.NewScanner([]byte(line))
			parser := d.ypayeedirectives(db)
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
			case "DRTV_SHORTNOTE":
				d.addBlockComment(trimstr(nodes[0]))
			case "DRTV_PAYEE_ALIAS":
				aliasv := nodes[2].(*parsec.Terminal).Value
				d.dpayeealias = append(d.dpayeealias, aliasv)
			case "DRTV_PAYEE_UUID":
				uuidv := nodes[2].(*parsec.Terminal).Value
				d.dpayeeuuid = append(d.dpayeeuuid, uuidv)
			}
		}
		return len(block), nil

	case "apply", "alias", "assert", "bucket", "capture", "check", "comment",
		"define", "fixed", "include", "test", "end", "year":
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

func (d *Directive) ydefine(db *Datastore) parsec.Parser {
	return parsec.And(
		func(nodes []parsec.ParsecNode) parsec.ParsecNode {
			d.dtype = "define"
			return d
		},
		ytokDirtDefine,
	)
}

func (d *Directive) yfixed(db *Datastore) parsec.Parser {
	return parsec.And(
		func(nodes []parsec.ParsecNode) parsec.ParsecNode {
			d.dtype = "fixed"
			return d
		},
		ytokDirtFixed,
	)
}

func (d *Directive) yinclude(db *Datastore) parsec.Parser {
	return parsec.And(
		func(nodes []parsec.ParsecNode) parsec.ParsecNode {
			d.dtype = "include"
			d.includefile = nodes[1].(*parsec.Terminal).Value
			d.includefile = strings.Trim(d.includefile, " \t")
			return d
		},
		ytokDirtInclude, ytokValue,
	)
}

func (d *Directive) ypayee(db *Datastore) parsec.Parser {
	return parsec.And(
		func(nodes []parsec.ParsecNode) parsec.ParsecNode {
			d.dtype = "payee"
			d.dpayee = nodes[1].(*parsec.Terminal).Value
			return d
		},
		ytokPayee, ytokPayeestr,
	)
}

func (d *Directive) ytest(db *Datastore) parsec.Parser {
	return parsec.And(
		func(nodes []parsec.ParsecNode) parsec.ParsecNode {
			d.dtype = "test"
			return d
		},
		ytokDirtTest,
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
	ytype := parsec.And(nil, ytokType, ytokHardSpace, ytokValue)
	ydefault := parsec.And(nil, ytokDefault)
	yshortnote := parsec.And(nil, ytokDirectivenote)
	y := parsec.OrdChoice(
		Vector2scalar,
		ynote, yalias, ypayee, ycheck, yassert, yeval, ytype, ydefault,
		yshortnote,
	)
	return y
}

func (d *Directive) ycommoditydirectives(db *Datastore) parsec.Parser {
	ynote := parsec.And(nil, ytokNote, ytokHardSpace, ytokValue)
	yformat := parsec.And(nil, ytokFormat, ytokHardSpace, ytokValue)
	ynomarket := parsec.And(nil, ytokNomarket)
	ycurrency := parsec.And(nil, ytokCommCurrency)
	ydefault := parsec.And(nil, ytokDefault)
	y := parsec.OrdChoice(
		Vector2scalar, ynote, yformat, ynomarket, ycurrency, ydefault,
	)
	return y
}

func (d *Directive) ypayeedirectives(db *Datastore) parsec.Parser {
	yalias := parsec.And(nil, ytokPayeeAlias, ytokHardSpace, ytokValue)
	yuuid := parsec.And(nil, ytokPayeeUuid, ytokHardSpace, ytokValue)
	y := parsec.OrdChoice(Vector2scalar, yalias, yuuid)
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
		account := db.GetAccount(d.accname).(*Account)
		account.addAlias(d.aliasname)
		db.addAlias(d.aliasname, d.accname)
		return nil

	case "assert":
		return fmt.Errorf("assert directive not-implemented")

	case "bucket":
		db.setBalancingaccount(d.accname)
		return nil

	case "capture":
		db.addCapture(d.capture, d.accname)
		return nil

	case "check":
		return fmt.Errorf("check directive not-implemented")

	case "comment":
		return fmt.Errorf("comment directive not-implemented")

	case "commodity":
		return db.declare(d)

	case "define":
		return fmt.Errorf("define directive not-implemented")

	case "fixed":
		return fmt.Errorf("fixed directive not-implemented")

	case "include":
		return nil

	case "payee":
		return db.declare(d)

	case "test":
		return fmt.Errorf("test directive not-implemented")

	case "end":
		return db.clearRootaccount()

	case "year":
		return db.setYear(d.year)
	}
	panic("unreachable code")
}

func (d *Directive) Secondpass(db *Datastore) error {
	return nil
}

func (d *Directive) addAccounttype(typenames []string, acc []string) []string {
	if acc == nil {
		acc = []string{}
	}
	for _, typename := range typenames {
		typename = strings.ToLower(typename)
		if api.HasString(acc, typename) {
			continue
		}
		switch typename {
		case "income":
			implies := []string{"credit"}
			acc = d.addAccounttype(implies, append(acc, typename))
		case "expense":
			implies := []string{"debit"}
			acc = d.addAccounttype(implies, append(acc, typename))
		default:
			acc = append(acc, typename)
		}
	}
	return acc
}

func (d *Directive) addBlockComment(comment string) {
	if d.comments == nil {
		d.comments = make([]string, 0)
	}
	d.comments = append(d.comments, comment)
}
