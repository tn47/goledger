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
			[]string{"-f", "dates.ldg", "register"},
			"dates.register1.ref",
		},
		[]interface{}{
			[]string{"-f", "dates.ldg", "register", "Expenses"},
			"dates.register2.ref",
		},
		[]interface{}{
			[]string{"-f", "dates.ldg", "register", "Expenses:Sta"},
			"dates.register3.ref",
		},
	}
	for _, testcase := range testcases {
		ref := testdataFile(testcase[1].(string))
		args := testcase[0].([]string)
		cmd := exec.Command(LEDGEREXEC, args...)
		out, _ := cmd.CombinedOutput()
		//ioutil.WriteFile(testcase[1].(string), out, 0660)
		if bytes.Compare(out, ref) != 0 {
			t.Logf(strings.Join(args, " "))
			t.Logf("expected %s", ref)
			t.Errorf("got %s", out)
		}
	}
}

func TestDate7(t *testing.T) {
	testcases := [][]interface{}{
		[]interface{}{
			[]string{"-f", "date7.ldg", "register"},
			"date7.register.ref",
		},
	}
	for _, testcase := range testcases {
		ref := testdataFile(testcase[1].(string))
		args := testcase[0].([]string)
		cmd := exec.Command(LEDGEREXEC, args...)
		out, _ := cmd.CombinedOutput()
		//ioutil.WriteFile(testcase[1].(string), out, 0660)
		if bytes.Compare(out, ref) != 0 {
			t.Logf(strings.Join(args, " "))
			t.Logf("expected %s", ref)
			t.Errorf("got %s", out)
		}
	}
}

func TestDrewr3(t *testing.T) {
	//testcases := [][]interface{}{
	//	[]interface{}{
	//		[]string{"-f", "drewr3.ldg", "balance"},
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
	//		t.Errorf("got %q", out)
	//	}
	//}
}

func TestFirst(t *testing.T) {
	testcases := [][]interface{}{
		[]interface{}{
			[]string{"-f", "first.ldg", "balance"},
			"first.balance1.ref",
		},
		[]interface{}{
			[]string{"-f", "first.ldg", "balance", "Assets"},
			"first.balance2.ref",
		},
		[]interface{}{
			[]string{"-f", "first.ldg", "balance", "Expenses"},
			"first.balance3.ref",
		},
		[]interface{}{
			[]string{"-f", "first.ldg", "register"},
			"first.register1.ref",
		},
		[]interface{}{
			[]string{"-f", "first.ldg", "register", "Expens|Check"},
			"first.register2.ref",
		},
		[]interface{}{
			[]string{"-f", "first.ldg", "register", "@", "KFC"},
			"first.register3.ref",
		},
		[]interface{}{
			[]string{"-f", "first.ldg", "register", "@", "^KFC"},
			"first.register4.ref",
		},
		[]interface{}{
			[]string{"-f", "first.ldg", "register", "Assets"},
			"first.register5.ref",
		},
		[]interface{}{
			[]string{"-f", "first.ldg", "register", "Check"},
			"first.register6.ref",
		},
		[]interface{}{
			[]string{"-f", "first.ldg", "register", "Dinning"},
			"first.register7.ref",
		},
		[]interface{}{
			[]string{"-f", "first.ldg", "register", "Expenses"},
			"first.register8.ref",
		},
		[]interface{}{
			[]string{"-f", "first.ldg", "register", "Expenses:Din"},
			"first.register9.ref",
		},
		[]interface{}{
			[]string{"-f", "first.ldg", "register", "Expenses:Sta"},
			"first.register10.ref",
		},
		[]interface{}{
			[]string{"-f", "first.ldg", "register", "Expens|Check"},
			"first.register11.ref",
		},
		[]interface{}{
			[]string{"-f", "first.ldg", "register", "^Check"},
			"first.register12.ref",
		},
		[]interface{}{
			[]string{"-f", "first.ldg", "register", "^nses"},
			"first.register13.ref",
		},
		[]interface{}{
			[]string{"-f", "first.ldg", "register", "nses"},
			"first.register14.ref",
		},
	}
	for _, testcase := range testcases {
		ref := testdataFile(testcase[1].(string))
		cmd := exec.Command(LEDGEREXEC, testcase[0].([]string)...)
		out, _ := cmd.CombinedOutput()
		//ioutil.WriteFile(testcase[1].(string), out, 0660)
		if bytes.Compare(out, ref) != 0 {
			t.Logf("expected %s", ref)
			t.Errorf("got %s", out)
		}
	}
}

func TestReimburse(t *testing.T) {
	testcases := [][]interface{}{
		[]interface{}{
			[]string{"-f", "reimburse.ldg", "balance"},
			"reimburse.balance1.ref",
		},
		[]interface{}{
			[]string{"-f", "reimburse.ldg", "register"},
			"reimburse.register1.ref",
		},
	}
	for _, testcase := range testcases {
		ref := testdataFile(testcase[1].(string))
		cmd := exec.Command(LEDGEREXEC, testcase[0].([]string)...)
		out, _ := cmd.CombinedOutput()
		//ioutil.WriteFile(testcase[1].(string), out, 0660)
		if bytes.Compare(out, ref) != 0 {
			t.Logf("expected %s", ref)
			t.Errorf("got %s", out)
		}
	}
}

func TestSecond(t *testing.T) {
	testcases := [][]interface{}{
		[]interface{}{
			[]string{"-f", "second.ldg", "balance"},
			"second.balance1.ref",
		},
		[]interface{}{
			[]string{"-f", "second.ldg", "balance",
				"^Assets", "^Liabilities"},
			"second.balance2.ref",
		},
		[]interface{}{
			[]string{"-f", "second.ldg", "balance",
				"Assets", "Liabilities.*"},
			"second.balance3.ref",
		},
		[]interface{}{
			[]string{"-f", "second.ldg", "balance",
				"Assets", "Liabilities"},
			"second.balance4.ref",
		},
		[]interface{}{
			[]string{"-f", "second.ldg", "balance",
				"Assets", "Liabilities.*"},
			"second.balance5.ref",
		},
		[]interface{}{
			[]string{"-f", "second.ldg", "balance",
				"^Assets", "^Liabilities"},
			"second.balance6.ref",
		},
		[]interface{}{
			[]string{"-f", "second.ldg", "balance",
				"^assets", "^liabilities"},
			"second.balance7.ref",
		},
	}
	for _, testcase := range testcases {
		ref := testdataFile(testcase[1].(string))
		cmd := exec.Command(LEDGEREXEC, testcase[0].([]string)...)
		out, _ := cmd.CombinedOutput()
		//ioutil.WriteFile(testcase[1].(string), out, 0660)
		if bytes.Compare(out, ref) != 0 {
			t.Logf("expected %s", ref)
			t.Errorf("got %s", out)
		}
	}
}

func TestMixedComm1(t *testing.T) {
	testcases := [][]interface{}{
		[]interface{}{
			[]string{"-f", "mixedcomm1.ldg", "balance"},
			"mixedcomm1.balance.ref",
		},
		[]interface{}{
			[]string{"-f", "mixedcomm1.ldg", "register"},
			"mixedcomm1.register.ref",
		},
	}
	for _, testcase := range testcases {
		ref := testdataFile(testcase[1].(string))
		cmd := exec.Command(LEDGEREXEC, testcase[0].([]string)...)
		out, _ := cmd.CombinedOutput()
		//ioutil.WriteFile(testcase[1].(string), out, 0660)
		if bytes.Compare(out, ref) != 0 {
			t.Logf("expected %s", ref)
			t.Errorf("got %s", out)
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
