package main

import "fmt"

import "github.com/prataprc/goparsec"
import "github.com/prataprc/golog"
import "github.com/prataprc/goledger/dblentry"

func firstpass(db *dblentry.Datastore, journalfile string) bool {
	var node parsec.ParsecNode

	lines := readlines(journalfile)

	iterate := blockiterate(lines)
	block, eof, err := iterate()
	for eof == false {
		if err != nil {
			log.Errorf("%v\n", err)
			return false
		}
		scanner := block[0]
		ytrans := dblentry.NewTransaction().Yledger(db)
		yprice := dblentry.NewPrice().Yledger(db)
		ydirective := dblentry.NewDirective().Yledger(db)
		y := parsec.OrdChoice(nil, ytrans, yprice, ydirective)

		node, scanner = y(scanner)

		switch obj := node.(type) {
		case *dblentry.Transaction:
			if len(block[1:]) > 0 {
				obj.Yledgerblock(db, block[1:])
			}
			db.Apply(obj)

		case *dblentry.Price:
			db.Apply(obj)

		case *dblentry.Directive:
			if len(block[1:]) > 0 {
				obj.Yledgerblock(db, block[1:])
			}
			db.Apply(obj)
		}
		block, eof, err = iterate()
	}

	return true
}

func blockiterate(lines []string) func() ([]parsec.Scanner, bool, error) {
	row := 0

	parseblock := func() []parsec.Scanner {
		var bs []byte
		var scanner parsec.Scanner

		scanners := []parsec.Scanner{}
		for ; row < len(lines); row++ {
			scanner = parsec.NewScanner([]byte(lines[row]))
			if bs, scanner = scanner.SkipWS(); scanner.Endof() || len(bs) == 0 {
				return scanners
			}
			scanners = append(scanners, scanner)
		}
		return scanners
	}

	return func() ([]parsec.Scanner, bool, error) {
		var bs []byte
		var scanner parsec.Scanner

		scanners := []parsec.Scanner{}
		for ; row < len(lines); row++ {
			scanner = parsec.NewScanner([]byte(lines[row]))
			if bs, scanner = scanner.SkipWS(); scanner.Endof() { // emptyline
				continue

			} else if len(bs) == 0 { // begin block
				row++
				scanners = append(scanners, scanner)
				scanners = append(scanners, parseblock()...)
				return scanners, row >= len(lines), nil

			} else {
				fmsg := "must be at the begnning: %v column:%v"
				cursor := scanner.GetCursor()
				return nil, false, fmt.Errorf(fmsg, row+1, cursor)
			}
		}
		return scanners, true, nil
	}
}
