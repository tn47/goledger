package dblentry

import "fmt"
import "strings"
import "testing"

import "github.com/prataprc/goparsec"

var _ = fmt.Sprintf("dummy")

func TestYmdy(t *testing.T) {
	testcases := [][]interface{}{
		// format: string, output
		{"2014/01/01", "2014-01-01 00:00:00"},
		{" 2014/01/1", "2014-01-01 00:00:00"},
		{"2014/1/01 ", "2014-01-01 00:00:00"},
		{"2014/1/1", "2014-01-01 00:00:00"},
		{"14/01/01", "2014-01-01 00:00:00"},
		{"14-01-01", "2014-01-01 00:00:00"},
		{"2014.01.01", "2014-01-01 00:00:00"},
	}

	for _, icase := range testcases {
		t.Logf("input %v", icase[0])
		scanner := parsec.NewScanner([]byte(icase[0].(string)))
		node, _ := Ydate(2014)(scanner)
		if node == nil && icase[1].(string) == "" {
			continue
		}
		parts := strings.Split(fmt.Sprintf("%v", node), " ")[:2]
		out := strings.Join(parts, " ")
		if out != icase[1].(string) {
			t.Errorf("expected %q, got %q", icase[1], out)
		}
	}
}

func TestYhms(t *testing.T) {
	testcases := [][]interface{}{
		{"2014/01/01 02:01:3", "2014-01-01 02:01:03"},
		{"2014/1/01 2:01:03", "2014-01-01 02:01:03"},
		{"2014/01/1 2:1:3", "2014-01-01 02:01:03"},
		{"2014/1/1 02:01:03", "2014-01-01 02:01:03"},
		{"2014/01/01 02:1:03", "2014-01-01 02:01:03"},
		{"2014/01/1 02:1:3", "2014-01-01 02:01:03"},
		{"2014/01/01 2:01:3", "2014-01-01 02:01:03"},
	}

	for _, icase := range testcases {
		t.Logf("input %v", icase[0])
		scanner := parsec.NewScanner([]byte(icase[0].(string)))
		node, _ := Ydate(2014)(scanner)
		if node == nil && icase[1].(string) == "" {
			continue
		}
		parts := strings.Split(fmt.Sprintf("%v", node), " ")[:2]
		out := strings.Join(parts, " ")
		if out != icase[1].(string) {
			t.Errorf("expected %q, got %q", icase[1], out)
		}
	}
}
