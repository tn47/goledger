package api

import "time"

type Datastorer interface {
	GetAccount(name string) Accounter

	// After firstpass

	Accountnames() []string

	Balance(obj interface{}) Commoditiser

	Balances() []Commoditiser

	SubAccounts(name string) []Accounter

	Formatter
}

type Transactor interface {
	Date() time.Time

	Payee() string

	GetPostings() []Poster
}

type Poster interface {
	Commodity() Commoditiser

	Payee() string

	Account() Accounter
}

type Commoditiser interface {
	Amount() float64

	Name() string

	String() string
}

type Accounter interface {
	Name() string

	Balance(obj interface{}) Commoditiser

	Balances() []Commoditiser

	HasPosting() bool

	Formatter
}

type Reporter interface {
	Transaction(Datastorer, Transactor) error

	Posting(Datastorer, Transactor, Poster) error

	BubblePosting(Datastorer, Transactor, Poster, Accounter) error

	Render(db Datastorer, args []string)
}

type Formatter interface {
	FmtBalances(Datastorer, Transactor, Poster, Accounter) [][]string

	FmtRegister(Datastorer, Transactor, Poster, Accounter) [][]string
}
