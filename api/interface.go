package api

import "time"

// Datastorer maintains the state of all accounts. Accounts are maintained as
// singleton objects.
type Datastorer interface {
	// GetAccount() to get the singleton object, if a singleton is not already
	// created for `name`, GetAccount() will create a new singleton object for
	// that name, and return the same.
	GetAccount(name string) Accounter

	// GetCommodity() returns the commodity object by name or creates an new
	// one. Note that this commodity object serves as the blue print for
	// commodity instansiation.
	GetCommodity(name string) Commoditiser

	// Accountnames return list of all account names.
	Accountnames() []string

	// Commoditynames return list of all commodity names.
	Commoditynames() []string

	// Balance amount, for commodity specified by `obj`, after all transactions
	// are tallied between the accounts. Note that each accounts can exchange
	// with any number of commodities.
	// `obj` can either be an instance implementing Commoditiser or it can be
	// string calling out the commodity name.
	Balance(obj interface{}) Commoditiser

	// Balances return the list of all commodities' balance amounts after all
	// transactions are tallied between the accounts.
	Balances() []Commoditiser

	// Firstpass to parse journal file.
	Firstpass(obj interface{}) error

	// Secondpass to apply journal entries.
	Secondpass() error

	// Firstpassok means firstpass completed on the datastore.
	Firstpassok()

	// Secondpassok means secondpass completed on the datastore.
	Secondpassok()

	// Clone this instance and all its nested reference.
	Clone(Reporter) Datastorer

	// AggregateTotal containing ledger for each posting.
	AggregateTotal(Transactor, Poster) error

	// IsCommodityDeclared return true if commodity is pre-declared
	IsCommodityDeclared(name string) bool

	// IsAccountDeclared return true if account is pre-declared
	IsAccountDeclared(name string) bool

	// IsPayeeDeclared return true if payee is pre-declared
	IsPayeeDeclared(name string) bool

	Formatter
}

// Transactor encapsulates a single transaction between one or more creditors
// and one or more debtors.
type Transactor interface {
	// Date of transaction.
	Date() time.Time

	// Payee is the person or organization that receives money, a person or
	// organization that is paid money.
	Payee() string

	// GetPostings return list of all postings under this transaction.
	GetPostings() []Poster

	// Printlines return the original lines from which transaction was parsed.
	Printlines() []string
}

// Poster encapsulates a single posting within a transaction.
type Poster interface {
	// Commodity posted as credit or debit.
	Commodity() Commoditiser

	// Lotprice return the lot-price of this posting's commodity.
	Lotprice() Commoditiser

	// Costprice return the cost-price of this posting's commodity.
	Costprice() Commoditiser

	// Balanceprice return the balance-price of this posting's account.
	Balanceprice() Commoditiser

	// Payee for this posting, if unspecified in the posting, shall return the
	// transaction's payee.
	Payee() string

	// Account to which the commodity should be posted.
	Account() Accounter
}

// Commoditiser encapsulates a commodity.
type Commoditiser interface {
	// Name of the commodity, like: $, INR, Gold, Oil etc...
	Name() string

	// Notes return list of notes declared on this commodity.
	Notes() []string

	// Amount as in quantity, not necessarily as value.
	Amount() float64

	// Currency is true if commodity is of type bool.
	Currency() bool

	// ApplyAmount commodities
	ApplyAmount(other Commoditiser) error

	// BalanceEqual is equality between two commodity, which implies equality
	// in Name(), Amount() and Currency().
	BalanceEqual(Commoditiser) (bool, error)

	// MakeSimilar create a new instance of commodity simlar to this commodity
	MakeSimilar(amount float64) Commoditiser

	// Directive return the commodity details as directive declaration.
	Directive() string

	String() string
}

// Accounter encapsulates an account.
type Accounter interface {
	// Name of the account.
	Name() string

	// Notes return list of notes declared on this account.
	Notes() []string

	// Balance amount, for commodity specified by `obj`, after all postings
	// from all transactions are applied on to this account. Note that each
	// accounts can exchange with any number of commodities.
	// `obj` can either be an instance implementing Commoditiser or it can be
	// string calling out the commodity name.
	Balance(obj interface{}) Commoditiser

	// Balances return the list of all commodities' balance amounts after all
	// postings form all transactions are applied on to this account.
	Balances() []Commoditiser

	// HasPosting return true if this account has ever participated in a
	// transaction posting.
	HasPosting() bool

	// Directive return the account details as directive declaration.
	Directive() string

	Formatter
}

// Reporter encapsulates callbacks that can be used by report generating
// plugins.
type Reporter interface {
	Firstpass(Datastorer, Transactor, Poster) error

	Transaction(Datastorer, Transactor) error

	Posting(Datastorer, Transactor, Poster) error

	BubblePosting(Datastorer, Transactor, Poster, Accounter) error

	Render(args []string, db Datastorer)

	Clone() Reporter
}

// Formatter implements are uniform tabularized {row,column} formatting across
// all types under dblentry package.
type Formatter interface {
	// FmtBalances used for `balance` reporting.
	FmtBalances(Datastorer, Transactor, Poster, Accounter) [][]string

	// FmtRegister used for `register` reporting.
	FmtRegister(Datastorer, Transactor, Poster, Accounter) [][]string

	// FmtEquity used for `equity` reporting.
	FmtEquity(Datastorer, Transactor, Poster, Accounter) [][]string
}
