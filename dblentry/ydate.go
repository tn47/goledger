package dblentry

import "fmt"
import "time"
import "regexp"
import "strconv"
import "strings"

import "github.com/prataprc/goparsec"

var century int

func Ydate(year, month int, format string) parsec.Parser {
	parsers := []interface{}{}
	parts := strings.Split(format, " ")
	parsers = append(parsers, Ymdy(parts[0]))
	if len(parts) == 2 {
		parsers = append(parsers, parsec.Maybe(maybenode, Yhns(parts[1])))
	}
	return parsec.And(
		func(nodes []parsec.ParsecNode) parsec.ParsecNode {
			ymd := nodes[0].([]interface{})
			year, month, date := ymd[0].(int), ymd[1].(int), ymd[2].(int)
			hour, minute, second := 0, 0, 0
			if len(nodes) == 2 {
				hns := nodes[1].([]interface{})
				hour, minute, second = hns[0].(int), hns[1].(int), hns[2].(int)
			}
			tm := time.Date(
				year, time.Month(month), date,
				hour, minute, second, 0,
				time.Local, /*locale*/
			)
			return tm
		},
		parsers...,
	)
}

func Ymdy(format string) parsec.Parser {
	pattern := "([^%]*)?(%[mdyY])"
	regc, err := regexp.Compile(pattern)
	if err != nil {
		panic(fmt.Errorf("unable to parse %q: %v\n", pattern, err))
	}
	matches := regc.FindAllStringSubmatch(format, -1)
	parsers := []interface{}{}
	for _, match := range matches {
		if match[1] != "" {
			parsers = append(parsers, parsec.Token(match[1], "LIMIT"))
		}
		switch match[2] {
		case "%Y":
			parsers = append(parsers, parsec.Token(`[0-9]{4}`, "YEAR"))
		case "%y":
			parsers = append(parsers, parsec.Token(`[0-9]{2}`, "YEAR"))
		case "%m":
			parsers = append(parsers, parsec.Token(`[0-9]{1,2}`, "MONTH"))
		case "%d":
			parsers = append(parsers, parsec.Token(`[0-9]{1,2}`, "DATE"))
		default:
			panic("unreachable code")
		}
	}

	y := parsec.And(
		func(nodes []parsec.ParsecNode) parsec.ParsecNode {
			var year, month, date int
			var err error

			for _, node := range nodes {
				switch t := node.(*parsec.Terminal); t.Name {
				case "LIMIT":
					continue

				case "YEAR":
					year, err = strconv.Atoi(t.Value)
					if err != nil {
						fmt.Printf("invalid YEAR at %v\n", t.Position)
					}
					if year < 100 {
						year = (century * 100) + year
					}

				case "MONTH":
					month, err = strconv.Atoi(t.Value)
					if err != nil {
						fmt.Printf("invalid MONTH at %v\n", t.Position)
					}

				case "DATE":
					date, err = strconv.Atoi(t.Value)
					if err != nil {
						fmt.Printf("invalid DATE at %v\n", t.Position)
					}

				default:
					panic("unreachable code")
				}
			}
			return []interface{}{year, month, date}
		},
		parsers...,
	)
	return y
}

func Yhns(format string) parsec.Parser {
	pattern := "([^%]*)?(%[hns])"
	regc, err := regexp.Compile(pattern)
	if err != nil {
		panic(fmt.Errorf("unable to parse %q: %v\n", pattern, err))
	}
	matches := regc.FindAllStringSubmatch(format, -1)
	parsers := []interface{}{}
	for _, match := range matches {
		if match[1] != "" {
			parsers = append(parsers, parsec.Token(match[1], "LIMIT"))
		}
		switch match[2] {
		case "%h":
			parsers = append(parsers, parsec.Token(`[0-9]{1,2}`, "HOUR"))
		case "%n":
			parsers = append(parsers, parsec.Token(`[0-9]{1,2}`, "MINUTE"))
		case "%s":
			parsers = append(parsers, parsec.Token(`[0-9]{1,2}`, "SECOND"))
		default:
			panic("unreachable code")
		}
	}

	y := parsec.And(
		func(nodes []parsec.ParsecNode) parsec.ParsecNode {
			var hour, minute, second int
			var err error

			for _, node := range nodes {
				switch t := node.(*parsec.Terminal); t.Name {
				case "LIMIT":
					continue

				case "HOUR":
					hour, err = strconv.Atoi(t.Value)
					if err != nil {
						fmt.Printf("invalid HOUR at %v\n", t.Position)
					}

				case "MINUTE":
					minute, err = strconv.Atoi(t.Value)
					if err != nil {
						fmt.Printf("invalid MONTH at %v\n", t.Position)
					}

				case "SECOND":
					second, err = strconv.Atoi(t.Value)
					if err != nil {
						fmt.Printf("invalid DATE at %v\n", t.Position)
					}

				default:
					panic("unreachable code")
				}
			}
			return []interface{}{hour, minute, second}
		},
		parsers...,
	)
	return y
}

func init() {
	century_year, _, _ := time.Now().Date()
	century = century_year / 100
}
