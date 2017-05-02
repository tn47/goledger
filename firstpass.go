package main

import "fmt"
import "strings"

import "github.com/prataprc/goparsec"
import "github.com/prataprc/golog"
import "github.com/prataprc/goledger/dblentry"

func firstpass(db *dblentry.Datastore, journalfile string) bool {
	var node parsec.ParsecNode

	lines := readlines(journalfile)

	iterate := blockiterate(lines)
	block, eof, err := iterate()
	for len(block) > 0 {
		if err != nil {
			log.Errorf("%v\n", err)
			return false
		}

		log.Debugf("parsing block: %v\n", block[0])
		scanner := parsec.NewScanner([]byte(block[0]))
		ytrans := dblentry.NewTransaction().Yledger(db)
		yprice := dblentry.NewPrice().Yledger(db)
		ydirective := dblentry.NewDirective().Yledger(db)
		y := parsec.OrdChoice(
			dblentry.Vector2scalar,
			ytrans, yprice, ydirective,
		)

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
	if eof == false {
		log.Errorf("expected eof")
	}

	return true
}

func blockiterate(lines []string) func() ([]string, bool, error) {
	row := 0

	parseblock := func() []string {
		blocklines := []string{}
		for ; row < len(lines); row++ {
			line := lines[row]
			if (len(line) > 0) && (line[0] == ' ' || line[0] == '\t') {
				if line = strings.TrimLeft(line, " \t"); line == "" {
					break
				}
				blocklines = append(blocklines, line)
			} else {
				break
			}
		}
		return blocklines
	}

	return func() ([]string, bool, error) {
		blocklines := []string{}
		for ; row < len(lines); row++ {
			line := lines[row]
			if len(line) == 0 {
				continue
			}
			if line[0] == ' ' || line[0] == '\t' {
				line = strings.TrimLeft(line, " \t")
				if line == "" { // emptyline
					continue
				} else {
					fmsg := "must be at the begnning: row:%v column: 0"
					return nil, false, fmt.Errorf(fmsg, row+1)
				}

			} else { // begin block
				row++
				blocklines = append(blocklines, line)
				blocklines = append(blocklines, parseblock()...)
				return blocklines, row >= len(lines), nil
			}
		}
		return blocklines, true, nil
	}
}
