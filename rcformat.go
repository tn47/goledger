package main

import "fmt"

type RCformat struct {
	rows    [][]string
	padding string
}

func NewRCformat() *RCformat {
	rcf := &RCformat{rows: [][]string{}, padding: " "}
	return rcf
}

func (rcf *RCformat) readsettings() *RCformat {
	return rcf
}

func (rcf *RCformat) Addrow(columns ...string) *RCformat {
	row := []string{}
	for _, col := range columns {
		col = rcf.padding + col + rcf.padding
		row = append(row, col)
	}
	rcf.rows = append(rcf.rows, row)
	return rcf
}

func (rcf *RCformat) FitWidth(maxwidths []int) {
	for i, maxwidth := range maxwidths {
		maxwidth -= 2
		for j, row := range rcf.rows {
			if ncol := len(row[i]); ncol > (maxwidth * 2) {
				row[i] = "!!"
			} else if ncol > maxwidth {
				nscrap := ncol - maxwidth
				x := nscrap / 2
				row[i] = row[i][0:x] + ".." + row[i][x+nscrap:]
			}
			rcf.rows[j] = row
		}
	}
}

func (rcf *RCformat) String() string {
	return fmt.Sprintf("RCformat{%v}\n", len(rcf.rows))
}
