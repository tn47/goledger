package api

import "time"
import "testing"
import "fmt"

var _ = fmt.Sprintf("dummy")

func TestValidateDate(t *testing.T) {
	y, m, d, h, n, s := -100000, 1, 1, 1, 1, 1
	tm := time.Date(y, time.Month(m), d, h, n, s, 0, time.Local)
	if ValidateDate(tm, y, m, d, h, n, s) == false {
		t.Errorf("expected true %v", tm)
	}
	y, m, d, h, n, s = 2013, 13, 1, 1, 1, 1
	tm = time.Date(y, time.Month(m), d, h, n, s, 0, time.Local)
	if ValidateDate(tm, y, m, d, h, n, s) {
		t.Errorf("expected false, %v", tm)
	}
	y, m, d, h, n, s = 2013, 2, 29, 1, 1, 1
	tm = time.Date(y, time.Month(m), d, h, n, s, 0, time.Local)
	if ValidateDate(tm, y, m, d, h, n, s) {
		t.Errorf("expected false, %v", tm)
	}
	y, m, d, h, n, s = 2013, 3, 28, 25, 1, 1
	tm = time.Date(y, time.Month(m), d, h, n, s, 0, time.Local)
	if ValidateDate(tm, y, m, d, h, n, s) {
		t.Errorf("expected false, %v", tm)
	}
	y, m, d, h, n, s = 2013, 3, 28, 23, 60, 1
	tm = time.Date(y, time.Month(m), d, h, n, s, 0, time.Local)
	if ValidateDate(tm, y, m, d, h, n, s) {
		t.Errorf("expected false, %v", tm)
	}
	y, m, d, h, n, s = 2013, 3, 28, 23, 59, 60
	tm = time.Date(y, time.Month(m), d, h, n, s, 0, time.Local)
	if ValidateDate(tm, y, m, d, h, n, s) {
		t.Errorf("expected false, %v", tm)
	}
}
