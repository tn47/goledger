package main

import "fmt"
import "time"
import "regexp"
import "strconv"

import "github.com/prataprc/goparsec"

var century int

func Ydate(year, month int, format string) parsec.Parser {
	pattern := "([^%]*)?(%[mdeyYAajuw])"
	regc, err := regexp.Compile(pattern)
	if err != nil {
		panic(fmt.Errorf("unable to parse %q: %v\n", pattern, err))
	}
	matches := regc.FindAllStringSubmatch(format, -1)
	parsers := []interface{}{}
	for i, match := range matches {
		if match[1] != "" {
			name := fmt.Sprintf("DATEDELIMIT-%v", i)
			parsers = append(parsers, parsec.Token(match[1], name))
		}
		switch match[2] {
		case "%Y", "%y":
			parsers = append(parsers, parsec.Token(match[1], "DATEYEAR"))
		case "%m":
			parsers = append(parsers, parsec.Token(match[1], "DATEMONTH"))
		case "%d":
			parsers = append(parsers, parsec.Token(match[1], "DATEDATE"))
		case "%A": // weekday information
			// TODO: we may not need to parse this, redundant information
		default:
			panic("unreachable code")
		}
	}

	y := parsec.And(
		func(nodes []parsec.ParsecNode) parsec.ParsecNode {
			var date int
			var err error

			for _, node := range nodes {
				switch t := node.(*parsec.Terminal); t.Name {
				case "DATEYEAR":
					year, err = strconv.Atoi(t.Value)
					if err != nil {
						fmt.Printf("invalid DATEYEAR at %v\n", t.Position)
					}
					if year < 100 {
						year = century + year
					}

				case "DATEMONTH":
					month, err = strconv.Atoi(t.Value)
					if err != nil {
						fmt.Printf("invalid DATEMONTH at %v\n", t.Position)
					}

				case "DATEDATE":
					date, err = strconv.Atoi(t.Value)
					if err != nil {
						fmt.Printf("invalid DATEDATE at %v\n", t.Position)
					}
				default:
					panic("unreachable code")
				}
			}

			tm := time.Date(
				year, time.Month(month), date, 0, 0, 0, 0, nil, /*TODO: locale*/
			)
			return tm
		},
		parsers...,
	)

	return y
}

func init() {
	century_year, _, _ := time.Now().Date()
	century = century_year / 100
}
