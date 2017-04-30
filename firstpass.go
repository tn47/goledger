package main

import "github.com/prataprc/goparsec"

func firstpass(db *Datastore, journalfile string) {
	var node parsec.ParsecNode
	var bs []byte

	lines := readlines(journalfile)

nextblock:
	for lineno := 0; lineno < len(lines); {
		line := lines[lineno]
		scanner := parsec.NewScanner([]byte(line))
		if bs, scanner = scanner.SkipWS(); scanner.Endof() == true {
			continue nextblock
		}

		trans := NewTransaction()
		price := NewPrice()
		directive := NewDirective()
		y := parsec.OrdChoice(nil, trans.Y(db), price.Y(db), directive.Y(db))

		node, scanner = y(scanner)

		switch block := node.(type) {
		case *Transaction:
			lineno++
			for ; lineno < len(lines); lineno++ {
				line := lines[lineno]
				scanner := parsec.NewScanner([]byte(line))
				if bs, scanner = scanner.SkipWS(); len(bs) == 0 {
					continue nextblock
				}
				posting := NewPosting()
				node, scanner = posting.Y(db)(scanner)
				block.Apply(db, node)
			}
			db.Apply(block)

		case *Price:
			lineno++
			db.Apply(block)

		case *Directive:
			lineno++
			for ; lineno < len(lines); lineno++ {
				line := lines[lineno]
				scanner := parsec.NewScanner([]byte(line))
				if bs, scanner = scanner.SkipWS(); len(bs) == 0 {
					continue nextblock
				}
				parser := block.Yattr(db)
				if parser == nil {
					continue nextblock
				}
				node, scanner = parser(scanner)
				block.Apply(db, node)
			}
			db.Apply(block)
		}
	}
	return
}
