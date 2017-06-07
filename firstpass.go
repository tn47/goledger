package main

import "fmt"
import "strings"
import "io/ioutil"
import "runtime/debug"
import "path/filepath"

import "github.com/prataprc/goparsec"
import "github.com/prataprc/golog"
import "github.com/tn47/goledger/dblentry"
import "github.com/tn47/goledger/api"
import "github.com/tn47/goledger/reports"

func dofirstpass(
	reporter api.Reporter,
	db *dblentry.Datastore, journalfile string) (err error) {

	data, err := ioutil.ReadFile(journalfile)
	if err != nil {
		log.Errorf("%v\n", err)
		return err
	}
	if db.Hasjournal(data) {
		log.Debugf("%q already processed !!\n", journalfile)
		return nil
	}
	db.Addjournal(journalfile, data)

	var lineno int
	var block []string
	var eof bool

	defer func() {
		if r := recover(); r != nil {
			log.Errorf("panic at %q:%v : %v\n", journalfile, lineno, r)
			log.Errorf("\n%s", api.GetStacktrace(2, debug.Stack()))
			err = fmt.Errorf("%s", r)
		}
	}()

	log.Debugf("firstpass %v\n", journalfile)
	var node parsec.ParsecNode

	lines, err := readlines(journalfile)
	if err != nil {
		return err
	}

	iterate := blockiterate(lines)
	lineno, block, eof, err = iterate()
	for len(block) > 0 {
		lineno -= len(block)
		if err != nil {
			log.Errorf("iterate at %q:%v : %v\n", journalfile, lineno, err)
			return err
		}
		log.Debugf("parsing block: %v\n", block[0])

		node, lineno, err = parseentry(lineno, block, journalfile, db)
		if err != nil {
			log.Errorf("parsec at %q:%v : %v\n", journalfile, lineno, err)
			return err
		} else if node != nil {
			if err := db.Firstpass(node); err != nil {
				fmsg := "%T at %q:%v : %v\n"
				log.Errorf(fmsg, node, journalfile, lineno-len(block)+1, err)
				return err
			}
		}

		tryinclude(reporter, db, node, journalfile)
		lineno, block, eof, err = iterate()
	}
	if err != nil {
		log.Errorf("%q : %v\n", journalfile, err)
	} else if eof == false {
		log.Errorf("%q : expected eof\n", journalfile)
	}
	return nil
}

func parseentry(
	lineno int,
	block []string,
	journalfile string,
	db *dblentry.Datastore) (parsec.ParsecNode, int, error) {

	var err error
	var index int

	scanner := parsec.NewScanner([]byte(block[0]))
	ytrans := dblentry.NewTransaction(journalfile).Yledger(db)
	yprice := dblentry.NewPrice().Yledger(db)
	ydirective := dblentry.NewDirective().Yledger(db)
	ycomment := dblentry.NewComment().Yledger(db)
	y := parsec.OrdChoice(
		dblentry.Vector2scalar,
		ytrans, yprice, ydirective, ycomment,
	)
	node, _ := y(scanner)
	switch obj := node.(type) {
	case *dblentry.Transaction:
		obj.Addlines(block[0])
		payee := strings.Trim(obj.Payee(), " \t")
		if api.Options.Stitch && payee == reports.PayeeOpeningBalance {
			return nil, lineno, nil
		}
		index, err = obj.Yledgerblock(db, block[1:])
		lineno += 1 + index
		obj.SetLineno(lineno)
		obj.Addlines(block[1:]...)

	case *dblentry.Directive:
		index, err = obj.Yledgerblock(db, block[1:])
		lineno += 1 + index

	case error:
		err = obj

	case *dblentry.Comment, *dblentry.Price:
	}
	return node, lineno, err
}

func tryinclude(
	reporter api.Reporter, db *dblentry.Datastore, node parsec.ParsecNode,
	includedby string) bool {

	if node == nil {
		return false
	}

	d, ok := node.(*dblentry.Directive)
	if ok && d.Type() == "include" {
		journalfile := d.Includefile()
		journalfile = strings.Trim(journalfile, "/")
		journalfile = filepath.Join(filepath.Dir(includedby), journalfile)
		reporter.Startjournal(journalfile, true /*included*/)
		dofirstpass(reporter, db, journalfile)
		return true
	}
	return false
}

func blockiterate(lines []string) func() (int, []string, bool, error) {
	row := 0

	parseblock := func() []string {
		blocklines := []string{}
		for ; row < len(lines); row++ {
			line := lines[row]
			if (len(line) > 0) && (line[0] == ' ' || line[0] == '\t') {
				if line1 := strings.TrimLeft(line, " \t"); line1 == "" {
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
				line1 := strings.TrimLeft(line, " \t")
				if line1 == "" { // emptyline
					continue
				} else {
					fmsg := "must be at the beginning: row:%v column: 0"
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
