package main

import "fmt"
import "time"
import "regexp"

import "github.com/prataprc/goparsec"

func yperiod(year, month, day int) parsec.Parser {
	inspec := parsec.And(
		func(nodes []parsec.ParsecNode) parsec.ParsecNode {
			return nodes[1].([3]int)
		},
		parsec.Atom("in", ""), yspec(year, month, day),
	)

	return parsec.And(
		func(nodes []parsec.ParsecNode) parsec.ParsecNode {
			return nodes
		},
		parsec.Maybe(nil, yinterval()),
		parsec.Maybe(nil, inspec),
		parsec.Maybe(nil, yspec(year, month, day)),
		parsec.Maybe(nil, ybegindate(year, month, day)),
		parsec.Maybe(nil, yenddate(year, month, day)),
	)
}

func yinterval() parsec.Parser {
	yevery := parsec.Atom("every", "EVERY")
	yint := parsec.Int()
	return parsec.OrdChoice(
		nil,
		parsec.Atom("every day", "EVERYDAY"),
		parsec.Atom("every week", "EVERYWEEK"),
		parsec.Atom("every month", "EVERYMONTH"),
		parsec.Atom("every quarter", "EVERYQUARTER"),
		parsec.Atom("every year", "EVERYYEAR"),
		parsec.And(nil, yevery, yint, parsec.Atom("days", "DAYS")),
		parsec.And(nil, yevery, yint, parsec.Atom("weeks", "WEEKS")),
		parsec.And(nil, yevery, yint, parsec.Atom("months", "MONTHS")),
		parsec.And(nil, yevery, yint, parsec.Atom("quarters", "QUARTERS")),
		parsec.And(nil, yevery, yint, parsec.Atom("years", "YEARS")),
		parsec.Atom("daily", "DAILY"),
		parsec.Atom("weekly", "WEEKLY"),
		parsec.Atom("biweekly", "BIWEEKLY"),
		parsec.Atom("monthly", "MONTHLY"),
		parsec.Atom("bimonthly", "BIMONTHLY"),
		parsec.Atom("quarterly", "QUARTERLY"),
		parsec.Atom("yearly", "YEARLY"),
	)
}

func ybegindate(year, month, day int) parsec.Parser {
	return parsec.And(
		func(nodes []parsec.ParsecNode) parsec.ParsecNode {
			return nodes[1].([3]int)
		},
		parsec.Token("from|since", ""), yspec(year, month, day),
	)
}

func yenddate(year, month, day int) parsec.Parser {
	return parsec.And(
		func(nodes []parsec.ParsecNode) parsec.ParsecNode {
			return nodes[1].([3]int)
		},
		parsec.Token("to|until|till", ""), yspec(year, month, day),
	)
}

func yspec(year, month, day int) parsec.Parser {
	delimit := `[/.-]`
	pattyr := `([0-9]{4})`
	mnname := `Jan|jan|Feb|feb|Mar|mar|Apr|apr|` +
		`May|may|Jun|jun|Jul|jul|Aug|aug|Sep|sep|Oct|oct|` +
		`Nov|nov|Dec|dec`
	pattmn := `([0-9]{1,2}|` + mnname + `)`
	pattdt := `([0-9]{1,2})`

	pattern := pattyr + delimit + pattmn + delimit + pattdt
	ydate1 := parsec.And( // 2004/10/1
		func(nodes []parsec.ParsecNode) parsec.ParsecNode {
			t := nodes[0].(*parsec.Terminal)
			regc, _ := regexp.Compile(pattern)
			parts := regc.FindStringSubmatch(string(t.Value))
			year = lookupyear[parts[1]]
			month = lookupmonth[parts[2]]
			day = lookupdate[parts[3]]
			return [3]int{year, month, day}
		},
		parsec.Token(pattern, ""),
	)

	pattern = pattyr + delimit + pattmn
	ydate2 := parsec.And( // 2004/10
		func(nodes []parsec.ParsecNode) parsec.ParsecNode {
			t := nodes[0].(*parsec.Terminal)
			regc, _ := regexp.Compile(pattern)
			parts := regc.FindStringSubmatch(string(t.Value))
			year, month = lookupyear[parts[1]], lookupmonth[parts[2]]
			return [3]int{year, month, day}
		},
		parsec.Token(pattern, ""),
	)
	ydate3 := parsec.And( // 2004
		func(nodes []parsec.ParsecNode) parsec.ParsecNode {
			year = lookupyear[string(nodes[0].(*parsec.Terminal).Value)]
			return [3]int{year, month, day}
		},
		parsec.Token(pattyr, ""),
	)
	ydate4 := parsec.And( // oct
		func(nodes []parsec.ParsecNode) parsec.ParsecNode {
			month = lookupmonth[string(nodes[0].(*parsec.Terminal).Value)]
			return [3]int{year, month, day}
		},
		parsec.Token(mnname, ""),
	)
	ybyday := parsec.And(
		func(nodes []parsec.ParsecNode) parsec.ParsecNode {
			k := string(nodes[0].(*parsec.Terminal).Value)
			tm := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.Local)
			tm = tm.Add(lookupbyday[k])
			return [3]int{tm.Year(), int(tm.Month()), tm.Day()}
		},
		parsec.Token(
			"yest|yesterday|last day|today|this day|tomm|tommorow|next day", ""),
	)
	ybymonth := parsec.And(
		func(nodes []parsec.ParsecNode) parsec.ParsecNode {
			what := string(nodes[0].(*parsec.Terminal).Value)
			switch what {
			case "last month":
				month--
			case "next month":
				month++
			}
			if month < 1 {
				year, month = year-1, 12-month
			} else if month > 12 {
				year, month = year+1, month%12
			}
			return [3]int{year, month, 0}
		},
		parsec.Token("last month|this month|next month", ""),
	)
	ybyyear := parsec.And(
		func(nodes []parsec.ParsecNode) parsec.ParsecNode {
			what := string(nodes[0].(*parsec.Terminal).Value)
			switch what {
			case "last year":
				year--
			case "next year":
				year++
			}
			return [3]int{year, 0, 0}
		},
		parsec.Token("last year|this year|next year", ""),
	)
	ybyquarter := parsec.And(
		func(nodes []parsec.ParsecNode) parsec.ParsecNode {
			what := string(nodes[0].(*parsec.Terminal).Value)
			month--
			quarter := (month / 3) + 1
			switch what {
			case "last quarter":
				quarter = (month / 3)
			case "next quarter":
				quarter = (month / 3) + 2
			}
			switch quarter {
			case 0:
				year, month = year-1, 10
			case 5:
				year, month = year+1, 1
			default:
				month = ((quarter - 1) * 3) + 1
			}
			return [3]int{year, month, 0}
		},
		parsec.Token("last quarter|this quarter|next quarter", ""),
	)
	return parsec.OrdChoice(
		func(nodes []parsec.ParsecNode) parsec.ParsecNode {
			return nodes[0].([3]int)
		},
		ydate1, ydate2, ydate3, ydate4, ybyday, ybymonth, ybyyear, ybyquarter,
	)
}

var lookupyear = map[string]int{}
var lookupmonth = map[string]int{
	"Jan": 1, "Feb": 2, "Mar": 3, "Apr": 4, "May": 5, "Jun": 6,
	"Jul": 7, "Aug": 8, "Sep": 9, "Oct": 10, "Nov": 11, "Dec": 12,
	"jan": 1, "feb": 2, "mar": 3, "apr": 4, "may": 5, "jun": 6,
	"jul": 7, "aug": 8, "sep": 9, "oct": 10, "nov": 11, "dec": 12,
}
var lookupdate = map[string]int{}
var lookupbyday = map[string]time.Duration{
	"yest":      -(24 * time.Hour),
	"yesterday": -(24 * time.Hour),
	"last day":  -(24 * time.Hour),
	"today":     0 * time.Hour,
	"this day":  0 * time.Hour,
	"tomm":      24 * time.Hour,
	"tommorow":  24 * time.Hour,
	"next day":  24 * time.Hour,
}
var lookupquarter = map[int]int{
	1: 1, 2: 1, 3: 1, 4: 2, 5: 2, 6: 2, 7: 3, 8: 3, 9: 3, 10: 4, 11: 4, 12: 4,
}

func init() {
	// year lookup
	for i := 0; i < 10000; i++ {
		lookupyear[fmt.Sprintf("%v", i)] = i
	}
	for i := 0; i < 10000; i++ {
		lookupyear[fmt.Sprintf("%04v", i)] = i
	}
	// month lookup
	for i := 0; i < 12; i++ {
		lookupyear[fmt.Sprintf("%v", i)] = i
	}
	for i := 0; i < 12; i++ {
		lookupyear[fmt.Sprintf("%02v", i)] = i
	}
	// date lookup
	for i := 0; i < 31; i++ {
		lookupyear[fmt.Sprintf("%v", i)] = i
	}
	for i := 0; i < 31; i++ {
		lookupyear[fmt.Sprintf("%02v", i)] = i
	}
}
