package dblentry

import "fmt"
import "time"
import "regexp"
import "strings"
import "strconv"

import "github.com/prataprc/goparsec"
import "github.com/bnclabs/golog"
import "github.com/tn47/goledger/api"

//---- ledger parser

// Ydate return a parser-combinator that can parse date/time string.
func Ydate(year int) parsec.Parser {
	pattdelimit := `[/.-]`
	pattmonth := `[0-9]{1,2}|Jan|jan|Feb|feb|Mar|mar|Apr|apr|` +
		`May|may|Jun|jun|Jul|jul|Aug|aug|Sep|sep|Oct|oct|` +
		`Nov|nov|Dec|dec`
	pattd2 := `[0-9]{1,2}`
	pattern := fmt.Sprintf(
		"([0-9]{2,4}%v)?(%v)%v(%v)"+ // date
			"( (%v):(%v):(%v))?", // time
		pattdelimit, pattmonth, pattdelimit, pattd2,
		pattd2, pattd2, pattd2,
	)

	// parts 1:year, 2:month, 3:date, 4:time, 5:hour, 6:minute, 7:second
	regc, err := regexp.Compile(pattern)
	if err != nil {
		panic(err)
	}

	return parsec.And(
		func(nodes []parsec.ParsecNode) parsec.ParsecNode {
			text := string(nodes[0].(*parsec.Terminal).Value)
			parts := regc.FindStringSubmatch(text)
			if parts[1] != "" {
				yr, _ := strconv.Atoi(strings.Trim(parts[1], pattdelimit))
				if yr < 100 {
					yr = ((year / 100) * 100) + yr
				}
				year = yr
			}
			month := mon2index[parts[2]]
			date, _ := strconv.Atoi(parts[3])
			hour, min, sec := 0, 0, 0
			if parts[4] != "" {
				hour, _ = strconv.Atoi(parts[5])
				min, _ = strconv.Atoi(parts[6])
				sec, _ = strconv.Atoi(parts[7])
			}
			tm := time.Date(
				year, time.Month(month), date, hour, min, sec, 0,
				time.Local, /*locale*/
			)
			ok := api.ValidateDate(tm, year, month, date, hour, min, sec)
			if ok == false {
				fmsg := "invalid date %v/%v/%v %v:%v:%v"
				return fmt.Errorf(fmsg, year, month, date, hour, min, sec)
			}
			log.Debugf("Ydate: %v\n", tm)
			return tm
		},
		parsec.Token(pattern, "DATETIME"),
	)
}

var mon2index = map[string]int{
	"Jan": 1, "Feb": 2, "Mar": 3, "Apr": 4, "May": 5, "Jun": 6,
	"Jul": 7, "Aug": 8, "Sep": 9, "Oct": 10, "Nov": 11, "Dec": 12,
	"jan": 1, "feb": 2, "mar": 3, "apr": 4, "may": 5, "jun": 6,
	"jul": 7, "aug": 8, "sep": 9, "oct": 10, "nov": 11, "dec": 12,
	"01": 1, "02": 2, "03": 3, "04": 4, "05": 5, "06": 6,
	"07": 7, "08": 8, "09": 9, "10": 10, "11": 11, "12": 12,
	"1": 1, "2": 2, "3": 3, "4": 4, "5": 5, "6": 6,
	"7": 7, "8": 8, "9": 9,
}
