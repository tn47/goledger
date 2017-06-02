package api

import "regexp"
import "strings"
import "fmt"

import "github.com/prataprc/goparsec"

func MakeFilterexpr(args []string) string {
	nargs := make([]string, 0, len(args))
	for _, arg := range args {
		arg = strings.Trim(arg, " \t")
		if len(arg) == 0 {
			continue
		}

		var prefix, suffix string
		if arg[0] == '(' && arg[len(arg)-1] == ')' {
			prefix, suffix, arg = "(", ")", arg[1:len(arg)-1]
		} else if arg[0] == '(' {
			prefix, suffix, arg = "(", "", arg[1:]
		} else if arg[len(arg)-1] == ')' {
			prefix, suffix, arg = "", ")", arg[:len(arg)-1]
		} else {
			prefix, suffix, arg = "", "", arg
		}
		switch strings.ToLower(arg) {
		case "and", "or", "not":
			nargs = append(nargs, prefix+arg+suffix)
		default:
			if arg[0] == '"' {
				nargs = append(nargs, prefix+arg+suffix)
			} else {
				nargs = append(nargs, prefix+`"`+arg+`"`+suffix)
			}
		}
	}
	return strings.Join(nargs, " ")
}

// Grammar
//
// YRegex     -> String+
// YParanExpr -> "(" YExpr ")"
// YValue     -> YRegex | YParanExpr | YNot
// YNot		  -> not YExpr
// YOrkleene  -> (or YAnd)*
// YOr		  -> YAnd YOrkleene
// YAndkleene -> (and YValue)*
// YAnd       -> YValue YAndkleene
// YExpr	  -> YOr

var YExpr parsec.Parser
var YAnd parsec.Parser
var YValue parsec.Parser

func init() {
	// YRegex
	YRegex := parsec.Many(
		func(nodes []parsec.ParsecNode) parsec.ParsecNode {
			//fmt.Println("YRegex", nodes)
			s := nodes[0].(string)
			op1, err := newMatchexpr(s[1 : len(s)-1])
			if err != nil {
				return err
			}
			if len(nodes) > 1 {
				for _, node := range nodes[1:] {
					s := node.(string)
					op2, err := newMatchexpr(s[1 : len(s)-1])
					if err != nil {
						return err
					}
					op1 = newFilterexpr("or", []*Filterexpr{op1, op2})
				}
			}
			return op1
		},
		parsec.String(), nil,
	)
	// YParanExpr -> "(" YExpr ")"
	YParanExpr := parsec.And(
		func(nodes []parsec.ParsecNode) parsec.ParsecNode {
			//fmt.Println("YParanExpr", nodes)
			if op, ok := nodes[1].(*Filterexpr); ok {
				return op
			}
			return nodes[1].(error)
		},
		parsec.Atom("(", "OPENPARAN"), &YExpr, parsec.Atom(")", "CLOSEPARAN"),
	)
	// YNot -> not YExpr
	YNot := parsec.And(
		func(nodes []parsec.ParsecNode) parsec.ParsecNode {
			//fmt.Println("YNot", nodes)
			op1, ok := nodes[1].(*Filterexpr)
			if ok == false {
				return nodes[1].(error)
			}
			fe := newFilterexpr("not", []*Filterexpr{op1})
			return fe
		},
		parsec.Atom("not", "NOT"), &YExpr,
	)
	// YValue -> YRegex | "(" YExpr ")"
	YValue = parsec.OrdChoice(
		func(nodes []parsec.ParsecNode) parsec.ParsecNode {
			//fmt.Println("YValue", nodes)
			return nodes[0]
		},
		YRegex, YParanExpr, YNot,
	)

	// (or YAnd)*
	YOrkleene := parsec.Kleene(
		nil, parsec.And(nil, parsec.Atom("or", "OR"), &YAnd), nil,
	)
	// YOr -> YAnd (or YAnd)*
	foldor := func(nds []parsec.ParsecNode, op1 *Filterexpr) parsec.ParsecNode {
		for _, nd := range nds {
			ns := nd.([]parsec.ParsecNode)
			op2, ok := ns[1].(*Filterexpr)
			if ok == false {
				return ns[1].(error)
			}
			op1 = newFilterexpr("or", []*Filterexpr{op1, op2})
		}
		return op1
	}
	YOr := parsec.And(
		func(nodes []parsec.ParsecNode) parsec.ParsecNode {
			//fmt.Println("YOr", nodes)
			op1, ok := nodes[0].(*Filterexpr)
			if ok == false {
				return nodes[0].(error)
			} else if nds := nodes[1].([]parsec.ParsecNode); len(nds) > 0 {
				return foldor(nds, op1)
			}
			return op1
		},
		&YAnd, YOrkleene,
	)

	// (and YValue)*
	YAndkleene := parsec.Kleene(
		nil, parsec.And(nil, parsec.Atom("and", "AND"), &YValue), nil,
	)
	// YAnd -> value (and value)*
	foldand := func(nds []parsec.ParsecNode, op1 *Filterexpr) parsec.ParsecNode {
		for _, nd := range nds {
			ns := nd.([]parsec.ParsecNode)
			op2, ok := ns[1].(*Filterexpr)
			if ok == false {
				return ns[1].(error)
			}
			op1 = newFilterexpr("and", []*Filterexpr{op1, op2})
		}
		return op1
	}
	YAnd = parsec.And(
		func(nodes []parsec.ParsecNode) parsec.ParsecNode {
			//fmt.Println("YAnd", nodes)
			op1, ok := nodes[0].(*Filterexpr)
			if ok == false {
				return nodes[0].(error)
			} else if nds := nodes[1].([]parsec.ParsecNode); len(nds) > 0 {
				return foldand(nds, op1)
			}
			return op1
		},
		&YValue, YAndkleene,
	)

	YExpr = parsec.OrdChoice(
		func(nodes []parsec.ParsecNode) parsec.ParsecNode {
			//fmt.Println("YExpr", nodes)
			return nodes[0]
		},
		YOr,
	)
}

type Filterexpr struct {
	op       string
	operands []*Filterexpr
	// for match op
	pattern string
	regc    *regexp.Regexp
}

func newFilterexpr(op string, operands []*Filterexpr) *Filterexpr {
	return &Filterexpr{op: op, operands: operands}
}

func newMatchexpr(pattern string) (*Filterexpr, error) {
	var err error

	fe := &Filterexpr{op: "match", pattern: pattern}
	if pattern != "" {
		if fe.regc, err = regexp.Compile(pattern); err != nil {
			return nil, err
		}
	}
	return fe, nil
}

func (fe *Filterexpr) Match(name string) bool {
	switch fe.op {
	case "match":
		return fe.regc.MatchString(name)
	case "and":
		op1, op2 := fe.operands[0], fe.operands[1]
		return op1.Match(name) && op2.Match(name)
	case "or":
		op1, op2 := fe.operands[0], fe.operands[1]
		return op1.Match(name) || op2.Match(name)
	case "not":
		return !fe.operands[0].Match(name)
	}
	panic("impossible situation")
}

func (fe *Filterexpr) String() string {
	switch fe.op {
	case "match":
		return fmt.Sprintf(`"re:%v"`, fe.pattern)
	case "and":
		op1, op2 := fe.operands[0], fe.operands[1]
		return fmt.Sprintf(`(%v and %v)`, op1, op2)
	case "or":
		op1, op2 := fe.operands[0], fe.operands[1]
		return fmt.Sprintf(`(%v or %v)`, op1, op2)
	case "not":
		op := fe.operands[0]
		return fmt.Sprintf(`(not %v)`, op)
	}
	panic("impossible situation")
}
