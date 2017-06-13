package dblentry

import "fmt"
import "strings"

import "github.com/prataprc/goparsec"
import "github.com/tn47/goledger/api"

var accountTypes = []string{
	"credit", "debit", "creditbalance", "debitbalance",
	"income", "expense", "accrual",
}

// Account implements api.Accounter{} interface.
type Account struct {
	name       string
	virtual    bool
	balanced   bool
	hasposting bool
	de         *DoubleEntry
	// from account directive
	notes    []string
	aliases  []string
	payees   []string
	types    []string
	comments []string
}

// NewAccount create a new instance of Account{}.
func NewAccount(name string) *Account {
	acc := &Account{
		name:  name,
		de:    NewDoubleEntry(name),
		types: []string{},
	}
	return acc
}

//---- local accessors

func (acc *Account) setPosting() {
	acc.hasposting = true
}

func (acc *Account) isVirtual() bool {
	return acc.virtual
}

func (acc *Account) addNote(note string) *Account {
	if note != "" {
		acc.notes = append(acc.notes, note)
	}
	return acc
}

func (acc *Account) addAlias(alias string) *Account {
	if alias != "" {
		acc.aliases = append(acc.aliases, alias)
	}
	return acc
}

func (acc *Account) addPayee(payee string) *Account {
	if payee != "" {
		acc.payees = append(acc.payees, payee)
	}
	return acc
}

func (acc *Account) addComments(comments ...string) *Account {
	if comments != nil {
		acc.comments = append(acc.comments, comments...)
	}
	return acc
}

func (acc *Account) isUnknown() bool {
	if acc.name == "Unknown" || strings.HasSuffix(acc.name, ":Unknown") {
		return true
	}
	return false
}

//---- api.Accounter methods.

func (acc *Account) Name() string {
	return acc.name
}

func (acc *Account) Notes() []string {
	return acc.notes
}

func (acc *Account) Balance(obj interface{}) (balance api.Commoditiser) {
	return acc.de.Balance(obj)
}

func (acc *Account) Balances() []api.Commoditiser {
	return acc.de.Balances()
}

func (acc *Account) Debit(obj interface{}) (balance api.Commoditiser) {
	return acc.de.Debit(obj)
}

func (acc *Account) Debits() []api.Commoditiser {
	return acc.de.Debits()
}

func (acc *Account) Credit(obj interface{}) (balance api.Commoditiser) {
	return acc.de.Credit(obj)
}

func (acc *Account) Credits() []api.Commoditiser {
	return acc.de.Credits()
}

func (acc *Account) Balanced() bool {
	return acc.balanced
}

func (acc *Account) HasPosting() bool {
	return acc.hasposting
}

func (acc *Account) IsIncome() bool {
	return api.HasString(acc.types, "income")
}

func (acc *Account) IsExpense() bool {
	return api.HasString(acc.types, "expense")
}

func (acc *Account) String() string {
	return fmt.Sprintf("%v", acc.name)
}

//---- api.Reporter{} methods

func (acc *Account) FmtBalances(
	db api.Datastorer, trans api.Transactor, p api.Poster,
	_ api.Accounter) [][]string {

	if len(acc.Balances()) == 0 {
		return nil
	}

	rows := make([][]string, 0) // Date, Accountname, Balance
	for _, balance := range acc.Balances() {
		if balance.Amount() != 0 || acc.HasPosting() == false {
			rows = append(rows, []string{"", "", balance.String()})
		}
	}
	if len(rows) > 0 { // last row to include date and account name.
		lastrow := rows[len(rows)-1]
		date := trans.Date().Format("2006/Jan/02")
		lastrow[0], lastrow[1] = date, acc.Name()
	}
	return rows
}

func (acc *Account) FmtDCBalances(
	db api.Datastorer, trans api.Transactor, p api.Poster,
	_ api.Accounter) [][]string {

	if len(acc.Balances()) == 0 {
		return nil
	}

	rows := make([][]string, 0) // Date, Accountname, Dr, Cr, Balance
	for _, bal := range acc.Balances() {
		name := bal.Name()
		dr, cr := acc.Debit(name), acc.Credit(name)
		cols := []string{"", "", dr.String(), cr.String(), bal.String()}
		rows = append(rows, cols)
	}

	if len(rows) > 0 { // last row to include date and account name.
		lastrow := rows[len(rows)-1]
		date := trans.Date().Format("2006/Jan/02")
		lastrow[0], lastrow[1] = date, acc.Name()
	}
	return rows
}

func (acc *Account) FmtEquity(
	db api.Datastorer, trans api.Transactor, p api.Poster,
	_ api.Accounter) [][]string {

	if len(acc.Balances()) == 0 {
		return nil
	}

	var rows [][]string

	for _, balance := range acc.Balances() {
		if balance.Amount() != 0 {
			rows = append(rows, []string{"", acc.Name(), balance.String()})
		}
	}
	return rows
}

func (acc *Account) FmtRegister(
	db api.Datastorer, trans api.Transactor, p api.Poster,
	_ api.Accounter) [][]string {

	panic("not supported")
}

func (acc *Account) FmtPassbook(
	db api.Datastorer, trans api.Transactor, p api.Poster,
	_ api.Accounter) [][]string {

	rows := make([][]string, 0)
	for _, balance := range acc.Balances() {
		if balance.Amount() != 0 {
			rows = append(rows, []string{"", "", "", "", balance.String()})
		}
	}

	if len(rows) > 0 {
		comm := p.Commodity()
		cols := rows[len(rows)-1] // pick the last balance entry
		cols[0], cols[1] = trans.Date().Format("2006/Jan/02"), p.Payee()
		if comm.IsDebit() {
			cols[2] = comm.String()
		} else {
			cols[3] = comm.MakeSimilar(-comm.Amount()).String()
		}
	}
	return rows
}

func (acc *Account) Directive() string {
	lines := []string{fmt.Sprintf("account %v", acc.name)}
	for _, note := range acc.notes {
		lines = append(lines, fmt.Sprintf("    note  %v", note))
	}
	for _, alias := range acc.aliases {
		lines = append(lines, fmt.Sprintf("    alias  %v", alias))
	}
	for _, payee := range acc.payees {
		lines = append(lines, fmt.Sprintf("    payee  %v", payee))
	}
	if len(acc.types) > 0 { // only if account's type is defined.
		line := fmt.Sprintf("    type  %v", strings.Join(acc.types, ","))
		lines = append(lines, line)
	}
	for _, comment := range acc.comments {
		lines = append(lines, fmt.Sprintf("    %v", comment))
	}
	return strings.Join(lines, "\n")
}

//---- ledger parser

// Yledger return a parser-combinator that can parse an account name.
func (acc *Account) Yledger(db *Datastore) parsec.Parser {
	y := parsec.OrdChoice(
		func(nodes []parsec.ParsecNode) parsec.ParsecNode {
			t := nodes[0].(*parsec.Terminal)
			name := strings.Trim(string(t.Value), " \t")
			switch t.Name {
			case "FULLACCNM":
				acc.name = name
				acc.virtual, acc.balanced = false, true
				return acc
			case "VFULLACCNM":
				acc.name = name[1 : len(name)-1]
				acc.virtual, acc.balanced = true, false
				return acc
			case "BFULLACCNM":
				acc.name = name[1 : len(name)-1]
				acc.virtual, acc.balanced = true, true
				return acc
			}
			panic(fmt.Errorf("unreachable code: terminal(%q)", t.Name))
		},
		ytokAccname, ytokVaccname, ytokBaccname,
	)
	return y
}

// Ypostaccn return a parser-combinator that parses an account name in the
// context of a posting.
func (acc *Account) Ypostaccn(db *Datastore) parsec.Parser {
	yacconly := parsec.And(
		func(nodes []parsec.ParsecNode) parsec.ParsecNode {
			return nodes[0]
		},
		ytokAccname, parsec.Parser(parsec.End()),
	)
	y := parsec.OrdChoice(
		func(nodes []parsec.ParsecNode) parsec.ParsecNode {
			t := nodes[0].(*parsec.Terminal)
			name := strings.Trim(string(t.Value), " \t")
			switch t.Name {
			case "POSTACCN1":
				acc.name = name
				acc.virtual, acc.balanced = false, true
				return acc
			case "FULLACCNM":
				acc.name = name
				acc.virtual, acc.balanced = false, true
				return acc
			case "VFULLACCNM":
				acc.name = name[1 : len(name)-1]
				acc.virtual, acc.balanced = true, false
				return acc
			case "BFULLACCNM":
				acc.name = name[1 : len(name)-1]
				acc.virtual, acc.balanced = true, true
				return acc
			}
			panic(fmt.Errorf("unreachable code: terminal(%q)", t.Name))
		},
		ytokPostacc1, ytokVaccname, ytokBaccname, yacconly,
	)
	return y
}

//---- engine

func (acc *Account) Firstpass(
	db *Datastore, trans *Transaction, p *Posting) error {

	return nil
}

func (acc *Account) Secondpass(
	db *Datastore, trans *Transaction, p *Posting) error {

	if err := p.account.addBalance(p.commodity); err != nil {
		return err
	}

	bal := p.account.Balance(p.commodity.name)
	if p.balprice != nil {
		if balok, err := bal.BalanceEqual(p.balprice); err != nil {
			return err

		} else if balok == false {
			accname := p.account.name
			fmsg := "account(%v) should balance as %s, got %s"
			return fmt.Errorf(fmsg, accname, p.balprice.String(), bal.String())
		}
	}
	return nil
}

func (acc *Account) Clone(ndb *Datastore) *Account {
	nacc := *acc
	nacc.de = acc.de.Clone()
	return &nacc
}

func (acc *Account) addBalance(commodity *Commodity) error {
	if err := acc.de.AddBalance(commodity); err != nil {
		return err
	}
	balance := acc.de.Balance(commodity.name).(*Commodity)
	if err := acc.assert(commodity, balance); err != nil {
		return err
	}
	return nil
}

func (acc *Account) assert(comm, bal *Commodity) error {
	if comm.IsCredit() {
		if err := acc.assertcredit(); err != nil {
			return err
		}
	}
	if comm.IsDebit() {
		if err := acc.assertdebit(); err != nil {
			return err
		}
	}
	if bal.IsCredit() {
		if err := acc.assertcrb(); err != nil {
			return err
		}
	}
	if bal.IsDebit() {
		if err := acc.assertdrb(); err != nil {
			return err
		}
	}
	return nil
}

func (acc *Account) assertcredit() error {
	if api.HasString(acc.types, "credit") {
		return nil
	}
	if api.HasString(acc.types, "debit") {
		return fmt.Errorf("account %q cannot be source", acc.name)
	}
	return nil
}

func (acc *Account) assertdebit() error {
	if api.HasString(acc.types, "debit") {
		return nil
	}
	if api.HasString(acc.types, "credit") {
		return fmt.Errorf("account %q cannot be target", acc.name)
	}
	return nil
}

func (acc *Account) assertcrb() error {
	if api.HasString(acc.types, "creditbalance") {
		return nil
	}
	if api.HasString(acc.types, "debitbalance") {
		return fmt.Errorf("account %q cannot have credit balance", acc.name)
	}
	return nil
}

func (acc *Account) assertdrb() error {
	if api.HasString(acc.types, "debitbalance") {
		return nil
	}
	if api.HasString(acc.types, "creditbalance") {
		return fmt.Errorf("account %q cannot have debit balance", acc.name)
	}
	return nil
}

// FitAccountname for formatting.
func FitAccountname(name string, maxwidth int) string {
	if len(name) < maxwidth {
		return name
	}
	scraplen := maxwidth - len(name)
	names := []string{}
	for _, name := range SplitAccount(name) {
		if scraplen <= 0 {
			names = append(names, name)
		}
		if len(name[3:]) < scraplen {
			names = append(names, name[:3])
			scraplen -= len(name[3:])
			continue
		}
		names = append(names, name[:len(name)-scraplen])
		scraplen = 0
	}
	return JoinAccounts(names)
}

// SplitAccount into segments.
func SplitAccount(name string) []string {
	return strings.Split(strings.Trim(name, ":"), ":")
}

// JoinAccounts from segments.
func JoinAccounts(segments []string) string {
	return strings.Join(segments, ":")
}
