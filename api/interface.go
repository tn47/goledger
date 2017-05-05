package api

import "time"

type Datastorer interface {
	Balance(obj interface{}) Commoditiser

	Balances() []Commoditiser
}

type Transactor interface {
	Date() time.Time

	Description() string

	GetPostings() []Poster
}

type Poster interface {
	Commodity() Commoditiser

	Account() Accounter
}

type Commoditiser interface {
	Amount() float64

	String() string
}

type Accounter interface {
	Name() string

	Balance(obj interface{}) Commoditiser

	Balances() []Commoditiser
}

type Reporter interface {
	Transaction(Datastorer, Transactor) error

	Posting(Datastorer, Transactor, Poster, Accounter) error

	BubblePosting(Datastorer, Transactor, Poster, Accounter) error

	Render(args []string)
}
