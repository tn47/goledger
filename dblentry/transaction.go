package dblentry

import "time"
import "fmt"
import "strings"
import "sort"

import "github.com/prataprc/goparsec"
import "github.com/prataprc/golog"
import "github.com/prataprc/goledger/api"

type Transaction struct {
	date     time.Time
	edate    time.Time
	code     string
	postings []*Posting
	tags     []string
	metadata map[string]interface{}
	notes    []string
	lineno   int
}

func NewTransaction() *Transaction {
	trans := &Transaction{
		tags:     []string{},
		metadata: map[string]interface{}{},
		notes:    []string{},
	}
	return trans
}

//---- accessor

func (trans *Transaction) Date() time.Time {
	return trans.date
}

func (trans *Transaction) GetPostings() []api.Poster {
	postings := []api.Poster{}
	for _, p := range trans.postings {
		postings = append(postings, p)
	}
	return postings
}

func (trans *Transaction) Metadata(key string) interface{} {
	if value, ok := trans.metadata[key]; ok {
		return value
	}
	return nil
}

func (trans *Transaction) SetMetadata(key string, value interface{}) {
	trans.metadata[key] = value
}

func (trans *Transaction) Payee() string {
	payee := trans.Metadata("payee")
	if payee != nil {
		return payee.(string)
	}
	return ""
}

func (trans *Transaction) State() string {
	state := trans.Metadata("state")
	if state != nil {
		return state.(string)
	}
	return ""
}

func (trans *Transaction) SetLineno(lineno int) {
	trans.lineno = lineno
}

func (trans *Transaction) Lineno() int {
	return trans.lineno
}

//---- ledger parser

func (trans *Transaction) Yledger(db *Datastore) parsec.Parser {
	// DATE
	ydate := Ydate(db.Year())
	// [=EDATE]
	yedate := parsec.And(
		func(nodes []parsec.ParsecNode) parsec.ParsecNode {
			return nodes[1] // EDATE
		},
		ytok_equal,
		ydate,
	)

	y := parsec.And(
		func(nodes []parsec.ParsecNode) parsec.ParsecNode {
			trans.date = nodes[0].(time.Time)
			if edate, ok := nodes[1].(time.Time); ok {
				trans.edate = edate
			}
			if t, ok := nodes[2].(*parsec.Terminal); ok {
				trans.SetMetadata("state", prefix2state[[]rune(t.Value)[0]])
			}
			if t, ok := nodes[3].(*parsec.Terminal); ok {
				trans.code = string(t.Value[1 : len(t.Value)-1])
			}

			payee := string(nodes[4].(*parsec.Terminal).Value)
			trans.SetMetadata("payee", payee)

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
		parsec.Maybe(maybenode, ytok_prefix),
		parsec.Maybe(maybenode, ytok_code),
		ytok_payeestr,
		parsec.Maybe(maybenode, ytok_transnote),
	)
	return y
}

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

		case Transnote:
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

func (trans *Transaction) ShouldBalance() bool {
	for _, posting := range trans.postings {
		virtual := posting.account.Virtual()
		balanced := posting.account.Balanced()
		if virtual == true && balanced == false {
			return false
		} else if balanced == false {
			return false
		}
	}
	return true
}

func (trans *Transaction) Defaultposting(
	db *Datastore, defacc *Account, commodity *Commodity) *Posting {

	posting := NewPosting(trans)
	posting.account = defacc
	posting.commodity = commodity
	return posting
}

func (trans *Transaction) Endposting(postings []*Posting) (*Posting, error) {
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

func (trans *Transaction) Autobalance1(
	db *Datastore, defaccount *Account) (bool, error) {

	if len(trans.postings) == 0 {
		return false, fmt.Errorf("empty transaction")

	} else if len(trans.postings) == 1 && defaccount != nil {
		commodity := trans.postings[0].Costprice()
		posting := trans.Defaultposting(db, defaccount, commodity)
		posting.commodity.InverseAmount()
		trans.postings = append(trans.postings, posting)
		return true, nil

	} else if len(trans.postings) == 1 {
		return false, fmt.Errorf("unbalanced transaction")
	}

	unbcs, _ := trans.DoBalance()
	if len(unbcs) == 0 {
		return true, nil
	}

	tallypost, err := trans.Endposting(trans.postings)
	if err != nil {
		return false, err
	}
	if len(unbcs) == 1 && tallypost == nil {
		return false, fmt.Errorf("unbalanced transaction")
	} else if tallypost == nil {
		return true, nil
	}

	tallypost.commodity = unbcs[0]
	tallypost.commodity.InverseAmount()
	if len(unbcs) > 1 {
		account := tallypost.account
		for _, unbc := range unbcs[1:] {
			posting := trans.Defaultposting(db, account, unbc)
			posting.commodity.InverseAmount()
			trans.postings = append(trans.postings, posting)
		}
	}
	return true, nil
}

func (trans *Transaction) DoBalance() ([]*Commodity, bool) {
	unbalanced := map[string]*Commodity{}
	for _, posting := range trans.postings {
		if posting.commodity == nil {
			continue
		}
		commodity := posting.Costprice()
		unbc, ok := unbalanced[commodity.name]
		if ok {
			unbc.Add(commodity)
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

func (trans *Transaction) Firstpass(db *Datastore) error {
	if trans.ShouldBalance() {
		defaccount := db.GetAccount(db.blncingaccnt).(*Account)
		if ok, err := trans.Autobalance1(db, defaccount); err != nil {
			return err
		} else if ok == false {
			return fmt.Errorf("unbalanced transaction")
		}
		log.Debugf("transaction balanced\n")
	}

	for _, posting := range trans.postings {
		if err := posting.Firstpass(db, trans); err != nil {
			return err
		}
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
