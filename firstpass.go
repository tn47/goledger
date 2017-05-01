package main

import "fmt"

import "github.com/prataprc/goparsec"

func firstpass(db *Datastore, journalfile string) {
	var node parsec.ParsecNode

	lines := readlines(journalfile)

	iterate := blockiterate(lines)
	block, eof, err := iterate()
	for eof == false {
		if err != nil {
			panic(err)
		}
		scanner := block[0]
		trans := NewTransaction()
		price := NewPrice()
		directive := NewDirective()
		y := parsec.OrdChoice(nil, trans.Y(db), price.Y(db), directive.Y(db))

		node, scanner = y(scanner)

		if trans, ok := node.(*Transaction); ok {
			if len(block[1:]) > 0 {
				trans.Applyblock(db, block[1:])
			}
			db.Apply(trans)

		} else if price, ok := node.(*Price); ok {
			db.Apply(price)

		} else if directive, ok := node.(*Directive); ok {
			if len(block[1:]) > 0 {
				directive.Applyblock(db, block[1:])
			}
			db.Apply(directive)
		}
		block, eof, err = iterate()
	}

	return
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
			if bs, scanner = scanner.SkipWS(); scanner.Endof() {
				continue

			} else if len(bs) == 0 {
				row++
				scanners = append(scanners, scanner)
				scanners = append(scanners, parseblock()...)
				return scanners, row >= len(lines), nil

			} else {
				fmsg := "syntax in beginning of line line: %v column:%v"
				cursor := scanner.GetCursor()
				return nil, false, fmt.Errorf(fmsg, row+1, cursor)
			}
		}
		return scanners, true, nil
	}
}
