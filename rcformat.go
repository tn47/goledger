package main

import "fmt"
import "strings"

import s "github.com/prataprc/gosettings"

type RCformat struct {
	heads []string
	rows  [][]string
	// settings
	width      int
	htxtalign  string
	ctxtalign  []string
	marginleft int
	padding    string
}

func NewRCformat(heads []string, setts s.Settings) *RCformat {
	rcf := (&RCformat{
		heads:   heads,
		rows:    [][]string{},
		padding: " ",
	}).readsettings(heads, setts)
	return rcf
}

func (rcf *RCformat) readsettings(heads []string, setts s.Settings) *RCformat {
	defsetts := defaultRCsetts(heads...)
	setts = make(s.Settings).Mixin(defsetts, setts)
	rcf.width = int(setts.Int64("width"))
	rcf.htxtalign = setts.String("htxtalign")
	rcf.ctxtalign = setts.Strings("ctxtalign")
	rcf.marginleft = int(setts.Int64("marginleft"))
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

func (rcf *RCformat) RenderBalance() {
	maxwidths := []int{14, 40, 14}
	rcf.FitWidth(maxwidths)

	marginleft := ""
	for i := 0; i < rcf.marginleft; i++ {
		marginleft += " "
	}
	for _, cols := range rcf.rows {
		line := strings.Join(cols, "")
		fmt.Printf("%v%v\n", marginleft, line)
	}
}

func defaultRCsetts(heads ...string) s.Settings {
	setts := s.Settings{
		"width":      80,
		"htxtalign":  "left",
		"marginleft": 4,
	}
	ctxtalign := make([]string, len(heads))
	for i := 0; i < len(heads); i++ {
		ctxtalign = append(ctxtalign, "left")
	}
	setts["ctxtalign"] = ctxtalign
	return setts
}
