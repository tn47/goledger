package main

import "github.com/prataprc/goparsec"

func firstpass(db *Datastore, scanner parsec.Scanner) {
	var node parsec.ParsecNode

	for scanner.Endof() == false {
		trans := NewTransaction()
		price := NewPrice()
		directive := NewDirective()

		y := parsec.OrdChoice(nil, trans.Y(db), price.Y(db), directive.Y(db))
		node, scanner = y(scanner)
		switch val := node.(type) {
		case *Transaction:
			scanner = val.Parse(db, scanner)
			trans = NewTransaction()
		case *Price:
			price = NewPrice()
		case *Directive:
			switch val.dtype {
			case "account":
				scanner = val.Parseaccount(db, val.account, scanner)
			}
			directive = NewDirective()
		}
	}
	return
}
