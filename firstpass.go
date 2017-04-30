package main

import "github.com/prataprc/goparsec"

func firstpass(db *Datastore, scanner parsec.Scanner) {
	var node parsec.ParsecNode

	trans := NewTransaction(db)
	price := NewPrice(db)
	d_account := NewDirectiveAccount(db)

	y := parsec.OrdChoice(nil, trans.Y(), price.Y(), d_account.Y())

	for scanner.Endof() == false {
		node, scanner = y(scanner)
		switch node.(type) {
		case *Transaction:
			scanner = trans.Parse(scanner)
			trans = NewTransaction(db)
		case *Price:
			price = NewPrice(db)
		case *DirectiveAccount:
			scanner = d_account.Parse(scanner)
			d_account = NewDirectiveAccount(db)
		}
	}
	return
}
