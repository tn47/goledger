package main

import "fmt"
import "testing"

import "github.com/prataprc/goparsec"

var _ = fmt.Sprintf("dummy")

func TestYmdy(t *testing.T) {
	testcases := map[string][][]interface{}{
		// format: string, output
		"%d-%m-%y": [][]interface{}{
			[]interface{}{"01-01-14", "2014-01-01 00:00:00 +0530 IST"},
			[]interface{}{"1-01-14", "2014-01-01 00:00:00 +0530 IST"},
			[]interface{}{"01-1-14", "2014-01-01 00:00:00 +0530 IST"},
			[]interface{}{"1-1-14", "2014-01-01 00:00:00 +0530 IST"},
			[]interface{}{"1:1-14", ""},
			[]interface{}{"1-1:14", ""},
			[]interface{}{"1:1:14", ""},
			[]interface{}{"1:1:2014", ""},
		},
		"%Y/%m/%d": [][]interface{}{
			[]interface{}{"2014/01/01", "2014-01-01 00:00:00 +0530 IST"},
			[]interface{}{" 2014/01/1", "2014-01-01 00:00:00 +0530 IST"},
			[]interface{}{"2014/1/01 ", "2014-01-01 00:00:00 +0530 IST"},
			[]interface{}{"2014/1/1", "2014-01-01 00:00:00 +0530 IST"},
			[]interface{}{"14/01/01", ""},
			[]interface{}{"14/01/01", ""},
			[]interface{}{"2014:01:01", ""},
		},
	}

	for format, tcase := range testcases {
		t.Logf("format: %v", format)
		for _, icase := range tcase {
			t.Logf("input %v", icase[0])
			scanner := parsec.NewScanner([]byte(icase[0].(string)))
			node, _ := Ydate(0, 0, format)(scanner)
			if node == nil && icase[1].(string) == "" {
				continue
			}
			out := fmt.Sprintf("%v", node)
			if out != icase[1].(string) {
				t.Errorf("expected %q, got %q", icase[1], out)
			}
		}
	}
}

func TestYhms(t *testing.T) {
	testcases := map[string][][]interface{}{
		// format: string, output
		"%Y/%m/%d %h:%n:%s": [][]interface{}{
			[]interface{}{"2014/01/01 02:01:3", "2014-01-01 02:01:03 +0530 IST"},
			[]interface{}{"2014/1/01 2:01:03", "2014-01-01 02:01:03 +0530 IST"},
			[]interface{}{"2014/01/1 2:1:3", "2014-01-01 02:01:03 +0530 IST"},
			[]interface{}{"2014/1/1 02:01:03", "2014-01-01 02:01:03 +0530 IST"},
			[]interface{}{"2014/01/01 02:1:03", "2014-01-01 02:01:03 +0530 IST"},
			[]interface{}{"2014/01/1 02:1:3", "2014-01-01 02:01:03 +0530 IST"},
			[]interface{}{"2014/01/01 2:01:3", "2014-01-01 02:01:03 +0530 IST"},
		},
	}

	for format, tcase := range testcases {
		t.Logf("format: %v", format)
		for _, icase := range tcase {
			t.Logf("input %v", icase[0])
			scanner := parsec.NewScanner([]byte(icase[0].(string)))
			node, _ := Ydate(0, 0, format)(scanner)
			if node == nil && icase[1].(string) == "" {
				continue
			}
			out := fmt.Sprintf("%v", node)
			if out != icase[1].(string) {
				t.Errorf("expected %q, got %q", icase[1], out)
			}
		}
	}
}
