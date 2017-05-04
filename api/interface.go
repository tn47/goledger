package api

import "time"

type Datastorer interface {
	Balance() float64
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
}

type Accounter interface {
	Name() string

	Balance() float64
}

type Reporter interface {
	Transaction(Datastorer, Transactor)

	Posting(Datastorer, Transactor, Poster, Accounter)

	BubblePosting(Datastorer, Transactor, Poster, Accounter)

	Render(args []string)
}
