package main

import "github.com/prataprc/goparsec"

func firstpass(db *Datastore, journalfile string) {
	var node parsec.ParsecNode

	for _, line := range readlines(journalfile) {
		scanner := parsec.NewScanner([]byte(line))
		for scanner.Endof() == false {
			trans := NewTransaction()
			price := NewPrice()
			directive := NewDirective()
			y := parsec.OrdChoice(nil, trans.Y(db), price.Y(db), directive.Y(db))
			node, scanner = y(scanner)
			switch block := node.(type) {
			case *Transaction:
				scanner = block.Parse(db, scanner)
			case *Price:
			case *Directive:
				switch block.dtype {
				case "account":
					scanner = block.Parseaccount(db, block.account, scanner)
				case "apply": // don't expect sub-directives
				}
			}
		}
	}
	return
}
