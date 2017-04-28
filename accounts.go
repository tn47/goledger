package main

var inclusives = []string{
	"asset", "liability", "capital", "equity", "income", "expense",
}

type Account struct {
	name              string
	virtual, balanced bool
}

func NewAccount(name string) *Account {
	virtual, balanced := false, true

	ln := len(name)
	switch first := name[0]; first {
	case '(':
		name, virtual, balanced = name[1:ln-1], true, false
	case '[':
		name, virtual, balanced = name[1:ln-1], true, true
	}
	return &Account{name: name, virtual: virtual, balanced: balanced}
}
