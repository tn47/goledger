package api

import "fmt"
import "reflect"
import "testing"
import "github.com/prataprc/goparsec"

var _ = fmt.Sprintf("dummy")

func TestFilterExpr(t *testing.T) {
	accnames := []string{
		"Expenses", "Expenses:Chats", "Expenses:Dinning",
		"Income", "Income:Salary", "Income:Chats", "Income:Travel",
		"Asset", "Asset:FD",
		"Liability", "Liability:Loan",
	}
	testcases := [][]interface{}{
		[]interface{}{
			[]string{"Expenses"},
			[]string{"Expenses", "Expenses:Chats", "Expenses:Dinning"},
		},
		[]interface{}{
			[]string{"Expenses", "Asset"},
			[]string{"Expenses", "Expenses:Chats",
				"Expenses:Dinning", "Asset", "Asset:FD"},
		},
		[]interface{}{
			[]string{"Expenses", "Assets"},
			[]string{"Expenses", "Expenses:Chats", "Expenses:Dinning"},
		},
		[]interface{}{
			[]string{"Expenses", "or", "Asset"},
			[]string{"Expenses", "Expenses:Chats",
				"Expenses:Dinning", "Asset", "Asset:FD"},
		},
		[]interface{}{
			[]string{"Expenses", "and", "Dinning"},
			[]string{"Expenses:Dinning"},
		},
		[]interface{}{
			[]string{"Expenses", "and", "Chat", "or", "Travel"},
			[]string{"Expenses:Chats", "Income:Travel"},
		},
		[]interface{}{
			[]string{"Expenses", "and", "(Chat", "or", "Travel)"},
			[]string{"Expenses:Chats"},
		},
		[]interface{}{
			[]string{"Expenses", "and", "(Chat", "or", "Dinning)"},
			[]string{"Expenses:Chats", "Expenses:Dinning"},
		},
		[]interface{}{
			[]string{"Expenses", "or", "Income", "and", "Salary"},
			[]string{"Expenses", "Expenses:Chats", "Expenses:Dinning",
				"Income:Salary"},
		},
		[]interface{}{
			[]string{"Expenses", "or", "Income", "or", "Asset"},
			[]string{"Expenses", "Expenses:Chats", "Expenses:Dinning",
				"Income", "Income:Salary", "Income:Chats", "Income:Travel",
				"Asset", "Asset:FD"},
		},
		[]interface{}{
			[]string{"Expenses", "Income", "Asset"},
			[]string{"Expenses", "Expenses:Chats", "Expenses:Dinning",
				"Income", "Income:Salary", "Income:Chats", "Income:Travel",
				"Asset", "Asset:FD"},
		},
		[]interface{}{
			[]string{"Expenses", "and", "Chat", "and", "Dinning"},
			[]string{},
		},
		[]interface{}{
			[]string{"not", "Expenses"},
			[]string{"Income", "Income:Salary", "Income:Chats", "Income:Travel",
				"Asset", "Asset:FD",
				"Liability", "Liability:Loan"},
		},
		[]interface{}{
			[]string{"not", "Expenses", "and", "Dinning"},
			[]string{"Expenses", "Expenses:Chats",
				"Income", "Income:Salary", "Income:Chats", "Income:Travel",
				"Asset", "Asset:FD",
				"Liability", "Liability:Loan"},
		},
		[]interface{}{
			[]string{"not", "Expenses", "and", "not", "Dinning"},
			[]string{"Expenses:Dinning",
				"Income", "Income:Salary", "Income:Chats", "Income:Travel",
				"Asset", "Asset:FD",
				"Liability", "Liability:Loan"},
		},
		[]interface{}{
			[]string{"not", "Expenses", "and", "not", "Dinning", "and", "Assets"},
			[]string{"Income", "Income:Salary", "Income:Chats", "Income:Travel",
				"Asset", "Asset:FD", "Liability", "Liability:Loan"},
		},
		[]interface{}{
			[]string{"not", "Expenses", "and", `(not`, "Dinning)", "and", "As"},
			[]string{"Expenses", "Expenses:Chats", "Expenses:Dinning",
				"Income", "Income:Salary", "Income:Chats", "Income:Travel",
				"Asset", "Asset:FD", "Liability", "Liability:Loan"},
		},
		[]interface{}{
			[]string{"not", "Expenses", "and", `("not"`, "Dinning)", "and", "As"},
			[]string{"Expenses", "Expenses:Chats", "Expenses:Dinning",
				"Income", "Income:Salary", "Income:Chats", "Income:Travel",
				"Asset", "Asset:FD", "Liability", "Liability:Loan"},
		},
	}
	for _, tcase := range testcases {
		args := tcase[0].([]string)
		arg := MakeFilterexpr(args)
		scanner := parsec.NewScanner([]byte(arg))
		node, _ := YExpr(scanner)
		if err, ok := node.(error); ok {
			t.Error(err)
		}
		fe := node.(*Filterexpr)
		names := []string{}
		for _, name := range accnames {
			if fe.Match(name) {
				names = append(names, name)
			}
		}
		if reflect.DeepEqual(names, tcase[1]) == false {
			t.Logf("%v", tcase)
			t.Fatalf("expected %v, got %v", tcase[1], names)
		}
	}
}
