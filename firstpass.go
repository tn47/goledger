package main

import "github.com/prataprc/goparsec"

func firstpass(context Context, scanner parsec.Scanner) {
	var node parsec.ParsecNode

	trans := NewTransaction(context)
	price := NewPrice(context)
	d_account := NewDirectiveAccount(context)

	y := parsec.OrdChoice(nil, trans.Y(), price.Y(), d_account.Y())

	for scanner.Endof() == false {
		node, scanner = y(scanner)
		switch node.(type) {
		case *Transaction:
			trans.Parsepostings(scanner)
			trans = NewTransaction(context)
		case *Price:
			price = NewPrice(context)
		case *DirectiveAccount:
			d_account.Parsedirective(scanner)
			d_account = NewDirectiveAccount(context)
		}
	}
	return
}
