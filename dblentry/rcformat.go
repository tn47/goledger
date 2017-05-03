package dblentry

import s "github.com/prataprc/gosettings"

type RCformat struct {
	heads []string
	rows  [][]string
	// settings
	width       int
	htxtalign   string
	ctxtalign   []string
	marginleft  int
	marginright int
	padding     string
}

func NewRCformat(heads []string, setts s.Settings) *RCformat {
	rc := (&RCformat{
		heads: heads,
		rows:  [][]string{},
	}).readsettings(heads, setts)
	return rc
}

func (rc *RCformat) readsettings(heads []string, setts s.Settings) *RCformat {
	defsetts := defaultRCsetts(heads...)
	setts = make(s.Settings).Mixin(defsetts, setts)
	rc.width = int(setts.Int64("width"))
	rc.htxtalign = setts.String("htxtalign")
	rc.ctxtalign = setts.Strings("ctxtalign")
	rc.marginleft = int(setts.Int64("marginleft"))
	rc.marginright = int(setts.Int64("marginright"))
	return rc
}

func (rc *RCformat) Addrow(columns ...string) *RCformat {
	row := []string{}
	for _, col := range columns {
		col = rc.padding + col + rc.padding
		row = append(row, col)
	}
	rc.rows = append(rc.rows, row)
	return rc
}

func (rc *RCformat) FitWidth(maxwidths []int) {
	for i, maxwidth := range maxwidths {
		maxwidth -= 2
		for j, row := range rc.rows {
			if ncol := len(row[i]); ncol > (maxwidth * 2) {
				row[i] = "!!"
			} else if ncol > maxwidth {
				nscrap := ncol - maxwidth
				x := nscrap / 2
				row[i] = row[i][0:x] + ".." + row[i][x+nscrap:]
			}
			rc.rows[j] = row
		}
	}
}

func defaultRCsetts(heads ...string) s.Settings {
	setts := s.Settings{
		"width":       80,
		"htxtalign":   "left",
		"marginleft":  4,
		"marginright": 4,
	}
	ctxtalign := make([]string, len(heads))
	for i := 0; i < len(heads); i++ {
		ctxtalign = append(ctxtalign, "left")
	}
	setts["ctxtalign"] = ctxtalign
	return setts
}
