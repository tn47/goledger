package tests

import "os"
import "fmt"
import "strings"
import "testing"
import "bytes"
import "compress/gzip"
import "io/ioutil"
import "os/exec"

var _ = fmt.Sprintf("dummy")
var LEDGEREXEC = "../ledger"

func TestDates(t *testing.T) {
	testcases := [][]interface{}{
		[]interface{}{
			[]string{"-f", "dates.ldg", "balance"},
			"dates.balance.ref",
		},
		[]interface{}{
			[]string{"-f", "testdata/dates.ldg", "register"},
			"dates.register1.ref",
		},
		[]interface{}{
			[]string{"-f", "testdata/dates.ldg", "register", "Expenses"},
			"dates.register2.ref",
		},
		[]interface{}{
			[]string{"-f", "testdata/dates.ldg", "register", "Expenses:Sta"},
			"dates.register3.ref",
		},
	}
	for _, testcase := range testcases {
		ref := testdataFile(testcase[1].(string))
		cmd := exec.Command(LEDGEREXEC, testcase[0].([]string)...)
		out, _ := cmd.CombinedOutput()
		if bytes.Compare(out, ref) != 0 {
			t.Logf("expected %q", ref)
			t.Logf("got %q", out)
		}
	}
}

func TestDrewr3(t *testing.T) {
	//testcases := [][]interface{}{
	//	[]interface{}{
	//		[]string{"-f", "testdata/drewr3.ldg", "balance"},
	//		"drewr3.balance.ref",
	//	},
	//}
	//for _, testcase := range testcases {
	//	ref := testdataFile(testcase[1].(string))
	//	cmd := exec.Command(LEDGEREXEC, testcase[0].([]string)...)
	//	out, _ := cmd.CombinedOutput()
	//	fmt.Println(testcase[0], "................")
	//	fmt.Println(string(ref))
	//	if bytes.Compare(out, ref) != 0 {
	//		t.Logf("expected %q", ref)
	//		t.Logf("got %q", out)
	//	}
	//}
}

func TestFirst(t *testing.T) {
	testcases := [][]interface{}{
		[]interface{}{
			[]string{"-f", "testdata/first.ldg", "balance"},
			"first.balance1.ref",
		},
		[]interface{}{
			[]string{"-f", "testdata/first.ldg", "balance", "Assets"},
			"first.balance2.ref",
		},
		[]interface{}{
			[]string{"-f", "testdata/first.ldg", "balance", "Expenses"},
			"first.balance3.ref",
		},
		[]interface{}{
			[]string{"-f", "testdata/first.ldg", "register"},
			"first.register1.ref",
		},
		[]interface{}{
			[]string{"-f", "testdata/first.ldg", "register", "Expens|Check"},
			"first.register2.ref",
		},
		[]interface{}{
			[]string{"-f", "testdata/first.ldg", "register", "@", "KFC"},
			"first.register3.ref",
		},
		[]interface{}{
			[]string{"-f", "testdata/first.ldg", "register", "@", "^KFC"},
			"first.register4.ref",
		},
		[]interface{}{
			[]string{"-f", "testdata/first.ldg", "register", "Assets"},
			"first.register5.ref",
		},
		[]interface{}{
			[]string{"-f", "testdata/first.ldg", "register", "Check"},
			"first.register6.ref",
		},
		[]interface{}{
			[]string{"-f", "testdata/first.ldg", "register", "Dinning"},
			"first.register7.ref",
		},
		[]interface{}{
			[]string{"-f", "testdata/first.ldg", "register", "Expenses"},
			"first.register8.ref",
		},
		[]interface{}{
			[]string{"-f", "testdata/first.ldg", "register", "Expenses:Din"},
			"first.register9.ref",
		},
		[]interface{}{
			[]string{"-f", "testdata/first.ldg", "register", "Expenses:Sta"},
			"first.register10.ref",
		},
		[]interface{}{
			[]string{"-f", "testdata/first.ldg", "register", "Expens|Check"},
			"first.register11.ref",
		},
		[]interface{}{
			[]string{"-f", "testdata/first.ldg", "register", "^Check"},
			"first.register12.ref",
		},
		[]interface{}{
			[]string{"-f", "testdata/first.ldg", "register", "^nses"},
			"first.register13.ref",
		},
		[]interface{}{
			[]string{"-f", "testdata/first.ldg", "register", "nses"},
			"first.register14.ref",
		},
	}
	for _, testcase := range testcases {
		ref := testdataFile(testcase[1].(string))
		cmd := exec.Command(LEDGEREXEC, testcase[0].([]string)...)
		out, _ := cmd.CombinedOutput()
		if bytes.Compare(out, ref) != 0 {
			t.Logf("expected %q", ref)
			t.Logf("got %q", out)
		}
	}
}

func TestReimburse(t *testing.T) {
	testcases := [][]interface{}{
		[]interface{}{
			[]string{"-f", "testdata/reimburse.ldg", "balance"},
			"reimburse.balance1.ref",
		},
		[]interface{}{
			[]string{"-f", "testdata/reimburse.ldg", "register"},
			"reimburse.register1.ref",
		},
	}
	for _, testcase := range testcases {
		ref := testdataFile(testcase[1].(string))
		cmd := exec.Command(LEDGEREXEC, testcase[0].([]string)...)
		out, _ := cmd.CombinedOutput()
		if bytes.Compare(out, ref) != 0 {
			t.Logf("expected %q", ref)
			t.Logf("got %q", out)
		}
	}
}

func TestSecond(t *testing.T) {
	testcases := [][]interface{}{
		[]interface{}{
			[]string{"-f", "testdata/second.ldg", "balance"},
			"second.balance1.ref",
		},
		[]interface{}{
			[]string{"-f", "testdata/second.ldg", "balance",
				"^Assets", "^Liabilities"},
			"second.balance2.ref",
		},
		[]interface{}{
			[]string{"-f", "testdata/second.ldg", "balance",
				"Assets", "Liabilities.*"},
			"second.balance3.ref",
		},
		[]interface{}{
			[]string{"-f", "testdata/second.ldg", "balance",
				"Assets", "Liabilities"},
			"second.balance4.ref",
		},
		[]interface{}{
			[]string{"-f", "testdata/second.ldg", "balance",
				"Assets", "Liabilities.*"},
			"second.balance5.ref",
		},
		[]interface{}{
			[]string{"-f", "testdata/second.ldg", "balance",
				"^Assets", "^Liabilities"},
			"second.balance6.ref",
		},
		[]interface{}{
			[]string{"-f", "testdata/second.ldg", "balance",
				"^assets", "^liabilities"},
			"second.balance7.ref",
		},
	}
	for _, testcase := range testcases {
		ref := testdataFile(testcase[1].(string))
		cmd := exec.Command(LEDGEREXEC, testcase[0].([]string)...)
		out, _ := cmd.CombinedOutput()
		if bytes.Compare(out, ref) != 0 {
			t.Logf("expected %q", ref)
			t.Logf("got %q", out)
		}
	}
}

func testdataFile(filename string) []byte {
	f, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	var data []byte
	if strings.HasSuffix(filename, ".gz") {
		gz, err := gzip.NewReader(f)
		if err != nil {
			panic(err)
		}
		data, err = ioutil.ReadAll(gz)
		if err != nil {
			panic(err)
		}
	} else {
		data, err = ioutil.ReadAll(f)
		if err != nil {
			panic(err)
		}
	}
	return data
}
