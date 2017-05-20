package dblentry

import "time"
import "fmt"
import "strings"
import "sort"

import "github.com/prataprc/goparsec"
import "github.com/prataprc/golog"
import "github.com/tn47/goledger/api"

// Transaction instance for every transaction in the journal file.
type Transaction struct {
	// immutable after firstpass
	date     time.Time
	edate    time.Time
	code     string
	tags     []string
	metadata map[string]interface{}
	notes    []string
	lineno   int

	postings []*Posting
}

// NewTransaction create a new transaction object.
func NewTransaction() *Transaction {
	trans := &Transaction{
		tags:     []string{},
		metadata: map[string]interface{}{},
		notes:    []string{},
	}
	return trans
}

//---- local accessors

func (trans *Transaction) getMetadata(key string) interface{} {
	if value, ok := trans.metadata[key]; ok {
		return value
	}
	return nil
}

func (trans *Transaction) setMetadata(key string, value interface{}) {
	trans.metadata[key] = value
}

func (trans *Transaction) getState() string {
	state := trans.getMetadata("state")
	if state != nil {
		return state.(string)
	}
	return ""
}

//---- exported accessors

// SetLineno in journal file for this transaction.
func (trans *Transaction) SetLineno(lineno int) {
	trans.lineno = lineno
}

// Lineno get lineno in journal file for this transaction.
func (trans *Transaction) Lineno() int {
	return trans.lineno
}

//---- api.Transactor methods.

func (trans *Transaction) Date() time.Time {
	return trans.date
}

func (trans *Transaction) Payee() string {
	payee := trans.getMetadata("payee")
	if payee != nil {
		return payee.(string)
	}
	return ""
}

func (trans *Transaction) GetPostings() []api.Poster {
	postings := []api.Poster{}
	for _, p := range trans.postings {
		postings = append(postings, p)
	}
	return postings
}

//---- ledger parser

// Yledger return a parser-combinator that can parse first line of a
// transaction.
func (trans *Transaction) Yledger(db *Datastore) parsec.Parser {
	// DATE
	ydate := Ydate(db.getYear())
	// [=EDATE]
	yedate := parsec.And(
		func(nodes []parsec.ParsecNode) parsec.ParsecNode {
			return nodes[1] // EDATE
		},
		ytokEqual,
		ydate,
	)

	y := parsec.And(
		func(nodes []parsec.ParsecNode) parsec.ParsecNode {
			trans.date = nodes[0].(time.Time)
			if edate, ok := nodes[1].(time.Time); ok {
				trans.edate = edate
			}
			if t, ok := nodes[2].(*parsec.Terminal); ok {
				trans.setMetadata("state", prefix2state[[]rune(t.Value)[0]])
			}
			if t, ok := nodes[3].(*parsec.Terminal); ok {
				trans.code = string(t.Value[1 : len(t.Value)-1])
			}

			payee := string(nodes[4].(*parsec.Terminal).Value)
			trans.setMetadata("payee", payee)

			if t, ok := nodes[5].(*parsec.Terminal); ok {
				note := string(t.Value)[1:]
				trans.notes = append(trans.notes, note)
			}

			fmsg := "trans.yledger date:%v code:%v payee:%v\n"
			log.Debugf(fmsg, trans.date, trans.code, payee)
			return trans
		},
		ydate,
		parsec.Maybe(maybenode, yedate),
		parsec.Maybe(maybenode, ytokPrefix),
		parsec.Maybe(maybenode, ytokCode),
		ytokPayeestr,
		parsec.Maybe(maybenode, ytokTransnote),
	)
	return y
}

// Yledgerblock return a parser combinaty that can parse all the posting
// within the transaction.
func (trans *Transaction) Yledgerblock(db *Datastore, block []string) error {
	if len(block) == 0 {
		return nil
	}

	var node parsec.ParsecNode

	for _, line := range block {
		scanner := parsec.NewScanner([]byte(line))
		posting := NewPosting(trans)
		node, scanner = posting.Yledger(db)(scanner)
		switch val := node.(type) {
		case *Posting:
			trans.postings = append(trans.postings, val)

		case *Tags:
			trans.tags = append(trans.tags, val.tags...)
			for k, v := range val.tagm {
				trans.metadata[k] = v
			}

		case typeTransnote:
			trans.notes = append(trans.notes, string(val))

		case error:
			return val
		}
		if scanner.Endof() == false {
			return fmt.Errorf("unable to parse posting")
		}
	}
	return nil
}

//---- engine

func (trans *Transaction) Firstpass(db *Datastore) error {
	for _, posting := range trans.postings {
		if err := posting.Firstpass(db, trans); err != nil {
			return err
		}
	}

	if trans.shouldBalance() {
		defaccount := db.GetAccount(db.getBalancingaccount()).(*Account)
		if ok, err := trans.autobalance1(db, defaccount); err != nil {
			return err
		} else if ok == false {
			return fmt.Errorf("unbalanced transaction")
		}
		log.Debugf("transaction balanced\n")
	}
	return nil
}

func (trans *Transaction) Secondpass(db *Datastore) error {
	for _, posting := range trans.postings {
		if err := posting.Secondpass(db, trans); err != nil {
			return fmt.Errorf("lineno: %v; %v", trans.lineno, err)
		}
	}
	return db.reporter.Transaction(db, trans)
}

func (trans *Transaction) Clone(ndb *Datastore) *Transaction {
	ntrans := *trans
	ntrans.postings = []*Posting{}
	for _, posting := range trans.postings {
		ntrans.postings = append(ntrans.postings, posting.Clone(ndb, &ntrans))
	}
	return &ntrans
}

func (trans *Transaction) shouldBalance() bool {
	for _, posting := range trans.postings {
		virtual := posting.isVirtual()
		balanced := posting.isBalanced()
		if virtual == true && balanced == false {
			return false
		} else if balanced == false {
			return false
		}
	}
	return true
}

func (trans *Transaction) defaultposting(
	db *Datastore, defacc *Account, commodity *Commodity) *Posting {

	posting := NewPosting(trans)
	posting.account = defacc
	posting.commodity = commodity
	return posting
}

func (trans *Transaction) endposting(postings []*Posting) (*Posting, error) {
	var tallypost *Posting
	for _, posting := range postings {
		if posting.commodity == nil && tallypost != nil {
			err := fmt.Errorf("Only one null posting allowed per transaction")
			return nil, err
		} else if posting.commodity == nil {
			tallypost = posting
		}
	}
	return tallypost, nil
}

func (trans *Transaction) autobalance1(
	db *Datastore, defaccount *Account) (bool, error) {

	if len(trans.postings) == 0 {
		return false, fmt.Errorf("empty transaction")

	} else if len(trans.postings) == 1 && defaccount != nil {
		commodity := trans.postings[0].getCostprice()
		posting := trans.defaultposting(db, defaccount, commodity)
		posting.commodity.doInverse()
		trans.postings = append(trans.postings, posting)
		return true, nil

	} else if len(trans.postings) == 1 {
		return false, fmt.Errorf("unbalanced transaction")
	}

	unbcs, _ := trans.doBalance()
	if len(unbcs) == 0 {
		return true, nil
	}

	tallypost, err := trans.endposting(trans.postings)
	if err != nil {
		return false, err
	}
	if len(unbcs) == 1 && tallypost == nil {
		return false, fmt.Errorf("unbalanced transaction")
	} else if tallypost == nil {
		return true, nil
	}

	tallypost.commodity = unbcs[0]
	tallypost.commodity.doInverse()
	if len(unbcs) > 1 {
		account := tallypost.account
		for _, unbc := range unbcs[1:] {
			posting := trans.defaultposting(db, account, unbc)
			posting.commodity.doInverse()
			trans.postings = append(trans.postings, posting)
		}
	}
	return true, nil
}

func (trans *Transaction) doBalance() ([]*Commodity, bool) {
	unbalanced := map[string]*Commodity{}
	for _, posting := range trans.postings {
		if posting.commodity == nil {
			continue
		}
		commodity := posting.getCostprice()
		unbc, ok := unbalanced[commodity.name]
		if ok {
			unbc.doAdd(commodity)
		} else {
			unbc = commodity
		}
		unbalanced[unbc.name] = unbc
	}
	commnames := []string{}
	for name := range unbalanced {
		commnames = append(commnames, name)
	}
	sort.Strings(commnames)

	unbcs := []*Commodity{}
	for _, name := range commnames {
		unbc := unbalanced[name]
		if unbc.amount != 0 {
			unbcs = append(unbcs, unbc)
		}
	}
	return unbcs, len(unbcs) > 1
}

// FitPayee for formatting.
func FitPayee(payee string, maxwidth int) string {
	if len(payee) < maxwidth {
		return payee
	}
	scraplen := len(payee) - maxwidth
	fields := []string{}
	for _, field := range strings.Fields(payee) {
		if scraplen <= 0 || len(field) <= 3 {
			fields = append(fields, field)
			continue
		}
		if len(field[3:]) < scraplen {
			fields = append(fields, field[:3])
			scraplen -= len(field[3:])
			continue
		}
		fields = append(fields, field[:len(field)-scraplen])
		scraplen = 0
	}
	return strings.Join(fields, " ")
}
