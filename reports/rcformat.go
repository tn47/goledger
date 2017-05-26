package reports

import "fmt"
import "strings"

import "github.com/prataprc/goparsec"
import "github.com/tn47/goledger/api"
import "github.com/tn47/goledger/dblentry"

// RCformat for {row, column} tabular formatting.
type RCformat struct {
	rows    [][]string
	padding string
}

// NewRCformat creates a new table of rows and colums.
func NewRCformat() *RCformat {
	rcf := &RCformat{rows: [][]string{}, padding: " "}
	return rcf
}

func (rcf *RCformat) readsettings() *RCformat {
	return rcf
}

func (rcf *RCformat) addrow(row ...string) *RCformat {
	rcf.rows = append(rcf.rows, row)
	return rcf
}

// FitAccountname for formatting
func (rcf *RCformat) FitAccountname(index, maxwidth int) int {
	for i, row := range rcf.rows {
		row[index] = dblentry.FitAccountname(row[index], maxwidth)
		rcf.rows[i] = row
	}
	return maxwidth
}

// FitPayee for formatting
func (rcf *RCformat) FitPayee(index, maxwidth int) int {
	for i, row := range rcf.rows {
		row[index] = dblentry.FitPayee(row[index], maxwidth)
		rcf.rows[i] = row
	}
	return maxwidth
}

func (rcf *RCformat) paddcells() {
	for y, row := range rcf.rows {
		for x, col := range row {
			row[x] = rcf.padding + col + rcf.padding
		}
		rcf.rows[y] = row
	}
}

// Fmsg format pattern for converting a row into report line.
func (rcf *RCformat) Fmsg(fmsg string) string {
	w := []interface{}{}
	for x := range rcf.rows[0] {
		w = append(w, rcf.maxwidth(rcf.column(x)))
	}
	return fmt.Sprintf(fmsg, w...)
}

func (rcf *RCformat) String() string {
	return fmt.Sprintf("RCformat{%v}\n", len(rcf.rows))
}

func (rcf *RCformat) maxwidth(col []string) int {
	if len(col) == 0 {
		return 0
	} else if len(col) == 1 {
		return len(col[0])
	}

	max := len(col[0])
	for _, s := range col[1:] {
		if len(s) > max {
			max = len(s)
		}
	}
	return max
}

func (rcf *RCformat) column(index int) []string {
	col := []string{}
	for _, row := range rcf.rows {
		col = append(col, row[index])
	}
	return col
}

func (rcf *RCformat) Clone() *RCformat {
	nrcf := *rcf
	nrcf.rows = [][]string{}
	return &nrcf
}

func CommodityColor(
	db api.Datastorer, comm *dblentry.Commodity, text string) (out interface{}) {

	if strings.Trim(text, " \t") == "" {
		return text
	}

	defer func() {
		if r := recover(); r != nil {
			out = text
		}
	}()

	scanner := parsec.NewScanner([]byte(text))
	node, _ := comm.Yledger(db.(*dblentry.Datastore))(scanner)
	if node.(api.Commoditiser).Amount() < 0 {
		return api.RedFn(text)
	}
	return text
}
