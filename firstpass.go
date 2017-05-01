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

		switch blockobj := node.(type) {
		case *Transaction:
			for _, scanner := range block[1:] {
				posting := NewPosting()
				node, scanner = posting.Y(db)(scanner)
				blockobj.Apply(db, node)
			}
			db.Apply(blockobj)

		case *Price:
			db.Apply(blockobj)

		case *Directive:
			for _, scanner := range block[1:] {
				parser := blockobj.Yattr(db)
				if parser == nil {
					continue
				}
				node, scanner = parser(scanner)
				blockobj.Apply(db, node)
			}
			db.Apply(blockobj)
		}
		block, eof, err = iterate()
	}

	return
}

func blockiterate(lines []string) func() ([]parsec.Scanner, bool, error) {
	var nextscanner parsec.Scanner

	row := 0

	blocklines := func() []parsec.Scanner {
		var bs []byte
		var scanner parsec.Scanner

		scanners := []parsec.Scanner{}
		for ; row < len(lines); row++ {
			scanner = parsec.NewScanner([]byte(lines[row]))
			bs, scanner = scanner.SkipWS()
			if len(bs) > 0 {
				if scanner.Endof() == false {
					scanners = append(scanners, scanner)
				}
				continue
			} else {
				row++
				return scanners
			}
		}
		return scanners
	}

	return func() ([]parsec.Scanner, bool, error) {
		var bs []byte
		var scanner parsec.Scanner

		scanners := []parsec.Scanner{}
		if nextscanner == nil {
			for ; row < len(lines); row++ {
				scanner = parsec.NewScanner([]byte(lines[row]))
				bs, scanner = scanner.SkipWS()
				if len(bs) == 0 && scanner.Endof() == false {
					scanners = append(scanners, scanner)
					scanners = append(scanners, blocklines()...)
					return scanners, row >= len(lines), nil
				} else if len(bs) > 0 {
					fmsg := "syntax in beginning of line line: %v column:%v"
					cursor := scanner.GetCursor()
					return nil, false, fmt.Errorf(fmsg, row+1, cursor)
				} else if scanner.Endof() {
					continue
				}
			}

		} else {
			scanners = append(scanners, nextscanner)
			scanners = append(scanners, blocklines()...)
			return scanners, row >= len(lines), nil
		}
		return scanners, true, nil
	}
}
