package main

import "fmt"
import "strings"

import "github.com/prataprc/goparsec"
import "github.com/prataprc/golog"
import "github.com/tn47/goledger/dblentry"

func firstpass(db *dblentry.Datastore, journalfile string) error {
	log.Debugf("firstpass %v\n", journalfile)
	var node parsec.ParsecNode

	lines := readlines(journalfile)

	iterate := blockiterate(lines)
	lineno, block, eof, err := iterate()
	for len(block) > 0 {
		if err != nil {
			log.Errorf("lineno %v: %v\n", lineno, err)
			return err
		}

		log.Debugf("parsing block: %v\n", block[0])
		scanner := parsec.NewScanner([]byte(block[0]))
		ytrans := dblentry.NewTransaction().Yledger(db)
		yprice := dblentry.NewPrice().Yledger(db)
		ydirective := dblentry.NewDirective().Yledger(db)
		ycomment := dblentry.NewComment().Yledger(db)
		y := parsec.OrdChoice(
			dblentry.Vector2scalar,
			ytrans, yprice, ydirective, ycomment,
		)

		node, _ = y(scanner)

		switch obj := node.(type) {
		case *dblentry.Transaction:
			err = obj.Yledgerblock(db, block[1:])
			obj.SetLineno(lineno - len(block))

		case *dblentry.Directive:
			err = obj.Yledgerblock(db, block[1:])

		case error:
			err = obj

		case *dblentry.Comment, *dblentry.Price:
		}
		if err != nil {
			log.Errorf("lineno %v: %v\n", lineno, err)
			return err
		}
		if err := db.Firstpass(node); err != nil {
			fmsg := "%T at lineno %v: %v\n"
			log.Errorf(fmsg, node, lineno-len(block), err)
			return err
		}
		lineno, block, eof, err = iterate()
	}
	if eof == false {
		log.Errorf("expected eof")
	}

	return nil
}

func blockiterate(lines []string) func() (int, []string, bool, error) {
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

	return func() (int, []string, bool, error) {
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
					return row + 1, nil, false, fmt.Errorf(fmsg, row+1)
				}

			} else { // begin block
				row++
				blocklines = append(blocklines, line)
				blocklines = append(blocklines, parseblock()...)
				return row + 1, blocklines, row >= len(lines), nil
			}
		}
		return row + 1, blocklines, true, nil
	}
}
