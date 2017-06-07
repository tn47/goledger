package tests

import "os"
import "fmt"
import "strings"
import "testing"
import "bytes"
import "io/ioutil"
import "os/exec"
import "compress/gzip"
import "path/filepath"

var _ = fmt.Sprintf("dummy")
var LEDGEREXEC = "../goledger"

var updateref = false

func TestCmdArgs(t *testing.T) {
	testcases := [][]interface{}{
		[]interface{}{
			[]string{"-f", "/a/b/xyz", "balance"},
			"refdata/cmdarg_f.ref",
		},
		[]interface{}{
			[]string{"-f", "first.ldg", "-o", "/a/b/xyz", "balance"},
			"refdata/cmdarg_o.ref",
		},
		[]interface{}{
			[]string{"-f", "first.ldg", "-fy", "abc", "balance"},
			"refdata/cmdarg_fy.ref",
		},
	}
	for _, testcase := range testcases {
		ref := testdataFile(testcase[1].(string))
		args := testcase[0].([]string)
		cmd := exec.Command(LEDGEREXEC, args...)
		out, _ := cmd.CombinedOutput()
		if updateref {
			ioutil.WriteFile(testcase[1].(string), out, 0660)
		}
		if bytes.Compare(out, ref) != 0 {
			t.Logf(strings.Join(args, " "))
			t.Logf("expected %s", ref)
			t.Errorf("got %s", out)
		}
	}
}

func TestBeginEnd(t *testing.T) {
	testcases := [][]interface{}{
		// balance
		[]interface{}{
			[]string{"-f", "beginend.ldg", "-begin", "2011/02/28", "-end",
				"2012/03/15", "balance"},
			"refdata/beginend.balance1.ref",
		},
		[]interface{}{
			[]string{"-f", "beginend.ldg", "-begin", "2011/02/28", "-end",
				"2012/03/16", "balance"},
			"refdata/beginend.balance2.ref",
		},
		[]interface{}{
			[]string{"-f", "beginend.ldg", "-begin", "2011/02/28", "balance"},
			"refdata/beginend.balance3.ref",
		},
		[]interface{}{
			[]string{"-f", "beginend.ldg", "-end", "2012/03/15", "balance"},
			"refdata/beginend.balance4.ref",
		},
		// register
		[]interface{}{
			[]string{"-f", "beginend.ldg", "-begin", "2011/02/28", "-end",
				"2012/03/15", "register"},
			"refdata/beginend.register1.ref",
		},
		[]interface{}{
			[]string{"-f", "beginend.ldg", "-begin", "2011/02/28", "-end",
				"2012/03/16", "register"},
			"refdata/beginend.register2.ref",
		},
		[]interface{}{
			[]string{"-f", "beginend.ldg", "-begin", "2012/03/15", "register"},
			"refdata/beginend.register3.ref",
		},
		[]interface{}{
			[]string{"-f", "beginend.ldg", "-end", "2012/03/15", "register"},
			"refdata/beginend.register4.ref",
		},
		// equity
		[]interface{}{
			[]string{"-f", "beginend.ldg", "-begin", "2011/02/28", "-end",
				"2012/03/15", "equity"},
			"refdata/beginend.equity1.ref",
		},
		[]interface{}{
			[]string{"-f", "beginend.ldg", "-begin", "2011/02/28", "-end",
				"2012/03/16", "equity"},
			"refdata/beginend.equity2.ref",
		},
		[]interface{}{
			[]string{"-f", "beginend.ldg", "-begin", "2012/03/15", "equity"},
			"refdata/beginend.equity3.ref",
		},
		[]interface{}{
			[]string{"-f", "beginend.ldg", "-end", "2012/03/15", "equity"},
			"refdata/beginend.equity4.ref",
		},
		// passbook
		[]interface{}{
			[]string{"-f", "beginend.ldg", "-begin", "2011/02/28", "-end",
				"2014/03/15", "passbook", "Assets:Checking"},
			"refdata/beginend.passbook1.ref",
		},
		[]interface{}{
			[]string{"-f", "beginend.ldg", "-begin", "2011/02/28", "-end",
				"2014/03/16", "passbook", "Assets:Checking"},
			"refdata/beginend.passbook2.ref",
		},
		[]interface{}{
			[]string{"-f", "beginend.ldg", "-begin", "2011/02/28", "passbook",
				"Assets:Checking"},
			"refdata/beginend.passbook3.ref",
		},
		[]interface{}{
			[]string{"-f", "beginend.ldg", "-end", "2014/03/15", "passbook",
				"Assets:Checking"},
			"refdata/beginend.passbook4.ref",
		},
	}
	for _, testcase := range testcases {
		ref := testdataFile(testcase[1].(string))
		args := testcase[0].([]string)
		cmd := exec.Command(LEDGEREXEC, args...)
		out, _ := cmd.CombinedOutput()
		if updateref {
			ioutil.WriteFile(testcase[1].(string), out, 0660)
		}
		if bytes.Compare(out, ref) != 0 {
			t.Logf(strings.Join(args, " "))
			t.Logf("expected %s", ref)
			t.Errorf("got %s", out)
		}
	}
}

func TestErrors(t *testing.T) {
	testcases := [][]interface{}{
		[]interface{}{
			[]string{"-f", "first.ldg", "list"},
			"refdata/first.listerr.ref",
		},
		[]interface{}{
			[]string{"-f", "error1.ldg", "balance"},
			"refdata/error1.ref",
		},
		[]interface{}{
			[]string{"-f", "error2.ldg", "-strict", "-checkpayee", "balance"},
			"refdata/error2.strict.ref",
		},
		[]interface{}{
			[]string{"-f", "dateerr1.ldg", "print"},
			"refdata/dateerr1.print.ref",
		},
		[]interface{}{
			[]string{"-f", "dateerr2.ldg", "print"},
			"refdata/dateerr2.print.ref",
		},
		[]interface{}{
			[]string{"-f", "dateerr3.ldg", "print"},
			"refdata/dateerr3.print.ref",
		},
	}
	for _, testcase := range testcases {
		ref := testdataFile(testcase[1].(string))
		args := testcase[0].([]string)
		cmd := exec.Command(LEDGEREXEC, args...)
		out, _ := cmd.CombinedOutput()
		if updateref {
			ioutil.WriteFile(testcase[1].(string), out, 0660)
		}
		if bytes.Compare(out, ref) != 0 {
			t.Logf(strings.Join(args, " "))
			t.Logf("expected %s", ref)
			t.Errorf("got %s", out)
		}
	}
}

func TestZeroSource(t *testing.T) {
	testcases := [][]interface{}{
		[]interface{}{
			[]string{"-f", "zerosource.ldg", "balance"},
			"refdata/zerosource.balance.ref",
		},
	}
	for _, testcase := range testcases {
		ref := testdataFile(testcase[1].(string))
		args := testcase[0].([]string)
		cmd := exec.Command(LEDGEREXEC, args...)
		out, _ := cmd.CombinedOutput()
		if updateref {
			ioutil.WriteFile(testcase[1].(string), out, 0660)
		}
		if bytes.Compare(out, ref) != 0 {
			t.Logf(strings.Join(args, " "))
			t.Logf("expected %s", ref)
			t.Errorf("got %s", out)
		}
	}
}
func TestAccountType(t *testing.T) {
	testcases := [][]interface{}{
		[]interface{}{
			[]string{"-f", "acctypeerr1.ldg", "balance"},
			"refdata/acctypeerr1.ref",
		},
		[]interface{}{
			[]string{"-f", "acctypeerr2.ldg", "balance"},
			"refdata/acctypeerr2.ref",
		},
		[]interface{}{
			[]string{"-f", "acctypeerr3.ldg", "balance"},
			"refdata/acctypeerr3.ref",
		},
	}
	for _, testcase := range testcases {
		ref := testdataFile(testcase[1].(string))
		args := testcase[0].([]string)
		cmd := exec.Command(LEDGEREXEC, args...)
		out, _ := cmd.CombinedOutput()
		if updateref {
			ioutil.WriteFile(testcase[1].(string), out, 0660)
		}
		if bytes.Compare(out, ref) != 0 {
			t.Logf(strings.Join(args, " "))
			t.Logf("expected %s", ref)
			t.Errorf("got %s", out)
		}
	}
}

func TestBasic(t *testing.T) {
	testcases := [][]interface{}{
		[]interface{}{
			[]string{"-f", "basic.ldg", "balance"},
			"refdata/basic.balance.ref",
		},
		[]interface{}{
			[]string{"-f", "basic.ldg", "-dc", "balance"},
			"refdata/basic.dcbalance.ref",
		},
		[]interface{}{
			[]string{"-f", "basic.ldg", "register"},
			"refdata/basic.register.ref",
		},
		[]interface{}{
			[]string{"-f", "basic.ldg", "equity"},
			"refdata/basic.equity.ref",
		},
		[]interface{}{
			[]string{"-f", "notes.ldg", "print"},
			"refdata/notes.print.ref",
		},
		[]interface{}{
			[]string{"-f", "lotdate.ldg", "print"},
			"refdata/lotdate.print.ref",
		},
	}
	for _, testcase := range testcases {
		ref := testdataFile(testcase[1].(string))
		args := testcase[0].([]string)
		cmd := exec.Command(LEDGEREXEC, args...)
		out, _ := cmd.CombinedOutput()
		if updateref {
			ioutil.WriteFile(testcase[1].(string), out, 0660)
		}
		if bytes.Compare(out, ref) != 0 {
			t.Logf(strings.Join(args, " "))
			t.Logf("expected %s", ref)
			t.Errorf("got %s", out)
		}
	}
}

func TestPassbook(t *testing.T) {
	testcases := [][]interface{}{
		[]interface{}{
			[]string{"-f", "balassert.ldg", "passbook", "Assets:Cash"},
			"refdata/balassert.passbook.ref",
		},
		[]interface{}{
			[]string{"-f", "date7.ldg", "passbook", "Assets:Checking"},
			"refdata/date7.passbook.ref",
		},
		[]interface{}{
			[]string{"-f", "dates.ldg", "passbook", "Assets:Checking"},
			"refdata/dates.passbook1.ref",
		},
		[]interface{}{
			[]string{"-f", "dates.ldg", "passbook", "Expenses:Dinning"},
			"refdata/dates.passbook2.ref",
		},
		[]interface{}{
			[]string{"-f", "elidingamount2.ldg", "passbook", "Assets:Cash"},
			"refdata/elidingamount2.passbook.ref",
		},
		[]interface{}{
			[]string{"-f", "fixprice.ldg", "passbook", "Assets:Checking"},
			"refdata/fixprice.passbook.ref",
		},
		[]interface{}{
			[]string{"-f", "mixedcomm1.ldg", "passbook", "EverQuest:Inventory"},
			"refdata/mixedcomm1.passbook.ref",
		},
	}
	for _, testcase := range testcases {
		ref := testdataFile(testcase[1].(string))
		args := testcase[0].([]string)
		cmd := exec.Command(LEDGEREXEC, args...)
		out, _ := cmd.CombinedOutput()
		if updateref {
			ioutil.WriteFile(testcase[1].(string), out, 0660)
		}
		if bytes.Compare(out, ref) != 0 {
			t.Logf(strings.Join(args, " "))
			t.Logf("expected %s", ref)
			t.Errorf("got %s", out)
		}
	}
}

func TestElidingAmount(t *testing.T) {
	testcases := [][]interface{}{
		[]interface{}{
			[]string{"-f", "elidingamount1.ldg", "balance"},
			"refdata/elidingamount1.balance.ref",
		},
		[]interface{}{
			[]string{"-f", "elidingamount1.ldg", "register"},
			"refdata/elidingamount1.register.ref",
		},
		[]interface{}{
			[]string{"-f", "elidingamount1.ldg", "equity"},
			"refdata/elidingamount1.equity.ref",
		},
		[]interface{}{
			[]string{"-f", "elidingamount2.ldg", "balance"},
			"refdata/elidingamount2.balance.ref",
		},
		[]interface{}{
			[]string{"-f", "elidingamount2.ldg", "register"},
			"refdata/elidingamount2.register.ref",
		},
		[]interface{}{
			[]string{"-f", "elidingamount2.ldg", "equity"},
			"refdata/elidingamount2.equity.ref",
		},
	}
	for _, testcase := range testcases {
		ref := testdataFile(testcase[1].(string))
		args := testcase[0].([]string)
		cmd := exec.Command(LEDGEREXEC, args...)
		out, _ := cmd.CombinedOutput()
		if updateref {
			ioutil.WriteFile(testcase[1].(string), out, 0660)
		}
		if bytes.Compare(out, ref) != 0 {
			t.Logf(strings.Join(args, " "))
			t.Logf("expected %s", ref)
			t.Errorf("got %s", out)
		}
	}
}

func TestAuxdate(t *testing.T) {
	testcases := [][]interface{}{
		[]interface{}{
			[]string{"-f", "auxdate.ldg", "balance"},
			"refdata/auxdate.balance.ref",
		},
		[]interface{}{
			[]string{"-f", "auxdate.ldg", "register"},
			"refdata/auxdate.register.ref",
		},
		[]interface{}{
			[]string{"-f", "auxdate.ldg", "equity"},
			"refdata/auxdate.equity.ref",
		},
	}
	for _, testcase := range testcases {
		ref := testdataFile(testcase[1].(string))
		args := testcase[0].([]string)
		cmd := exec.Command(LEDGEREXEC, args...)
		out, _ := cmd.CombinedOutput()
		if updateref {
			ioutil.WriteFile(testcase[1].(string), out, 0660)
		}
		if bytes.Compare(out, ref) != 0 {
			t.Logf(strings.Join(args, " "))
			t.Logf("expected %s", ref)
			t.Errorf("got %s", out)
		}
	}
}

func TestTranscode(t *testing.T) {
	testcases := [][]interface{}{
		[]interface{}{
			[]string{"-f", "transcode.ldg", "balance"},
			"refdata/transcode.balance.ref",
		},
		[]interface{}{
			[]string{"-f", "transcode.ldg", "register"},
			"refdata/transcode.register.ref",
		},
		[]interface{}{
			[]string{"-f", "transcode.ldg", "equity"},
			"refdata/transcode.equity.ref",
		},
	}
	for _, testcase := range testcases {
		ref := testdataFile(testcase[1].(string))
		args := testcase[0].([]string)
		cmd := exec.Command(LEDGEREXEC, args...)
		out, _ := cmd.CombinedOutput()
		if updateref {
			ioutil.WriteFile(testcase[1].(string), out, 0660)
		}
		if bytes.Compare(out, ref) != 0 {
			t.Logf(strings.Join(args, " "))
			t.Logf("expected %s", ref)
			t.Errorf("got %s", out)
		}
	}
}

func TestBalanceErr(t *testing.T) {
	testcases := [][]interface{}{
		[]interface{}{
			[]string{"-f", "balerr1.ldg", "balance"},
			"refdata/balerr1.ref",
		},
		[]interface{}{
			[]string{"-f", "balerr2.ldg", "balance"},
			"refdata/balerr2.ref",
		},
	}
	for _, testcase := range testcases {
		ref := testdataFile(testcase[1].(string))
		args := testcase[0].([]string)
		cmd := exec.Command(LEDGEREXEC, args...)
		out, _ := cmd.CombinedOutput()
		if updateref {
			ioutil.WriteFile(testcase[1].(string), out, 0660)
		}
		if bytes.Compare(out, ref) != 0 {
			t.Logf(strings.Join(args, " "))
			t.Logf("expected %s", ref)
			t.Errorf("got %s", out)
		}
	}
}

func TestBalanceAssert(t *testing.T) {
	testcases := [][]interface{}{
		[]interface{}{
			[]string{"-f", "balassert.ldg", "balance"},
			"refdata/balassert.balance.ref",
		},
		[]interface{}{
			[]string{"-f", "balassert.ldg", "register"},
			"refdata/balassert.register.ref",
		},
		[]interface{}{
			[]string{"-f", "balassert.ldg", "equity"},
			"refdata/balassert.equity.ref",
		},
	}
	for _, testcase := range testcases {
		ref := testdataFile(testcase[1].(string))
		args := testcase[0].([]string)
		cmd := exec.Command(LEDGEREXEC, args...)
		out, _ := cmd.CombinedOutput()
		if updateref {
			ioutil.WriteFile(testcase[1].(string), out, 0660)
		}
		if bytes.Compare(out, ref) != 0 {
			t.Logf(strings.Join(args, " "))
			t.Logf("expected %s", ref)
			t.Errorf("got %s", out)
		}
	}
}

func TestExplicitCost(t *testing.T) {
	testcases := [][]interface{}{
		[]interface{}{
			[]string{"-f", "explicitcost.ldg", "balance"},
			"refdata/explicitcost.balance.ref",
		},
		[]interface{}{
			[]string{"-f", "explicitcost.ldg", "-dc", "balance"},
			"refdata/explicitcost.dcbalance.ref",
		},
		[]interface{}{
			[]string{"-f", "explicitcost.ldg", "register"},
			"refdata/explicitcost.register.ref",
		},
		[]interface{}{
			[]string{"-f", "explicitcost.ldg", "equity"},
			"refdata/explicitcost.equity.ref",
		},
	}
	for _, testcase := range testcases {
		ref := testdataFile(testcase[1].(string))
		args := testcase[0].([]string)
		cmd := exec.Command(LEDGEREXEC, args...)
		out, _ := cmd.CombinedOutput()
		if updateref {
			ioutil.WriteFile(testcase[1].(string), out, 0660)
		}
		if bytes.Compare(out, ref) != 0 {
			t.Logf(strings.Join(args, " "))
			t.Logf("expected %s", ref)
			t.Errorf("got %s", out)
		}
	}
}

func TestTotalCost(t *testing.T) {
	testcases := [][]interface{}{
		[]interface{}{
			[]string{"-f", "totalcost.ldg", "balance"},
			"refdata/totalcost.balance.ref",
		},
		[]interface{}{
			[]string{"-f", "totalcost.ldg", "register"},
			"refdata/totalcost.register.ref",
		},
		[]interface{}{
			[]string{"-f", "totalcost.ldg", "equity"},
			"refdata/totalcost.equity.ref",
		},
	}
	for _, testcase := range testcases {
		ref := testdataFile(testcase[1].(string))
		args := testcase[0].([]string)
		cmd := exec.Command(LEDGEREXEC, args...)
		out, _ := cmd.CombinedOutput()
		if updateref {
			ioutil.WriteFile(testcase[1].(string), out, 0660)
		}
		if bytes.Compare(out, ref) != 0 {
			t.Logf(strings.Join(args, " "))
			t.Logf("expected %s", ref)
			t.Errorf("got %s", out)
		}
	}
}

func TestDates(t *testing.T) {
	testcases := [][]interface{}{
		[]interface{}{
			[]string{"-f", "dates.ldg", "balance"},
			"refdata/dates.balance.ref",
		},
		[]interface{}{
			[]string{"-f", "dates.ldg", "-dc", "balance"},
			"refdata/dates.dcbalance.ref",
		},
		[]interface{}{
			[]string{"-f", "dates.ldg", "-nosubtotal", "balance"},
			"refdata/dates.balance.nosubtotal.ref",
		},
		[]interface{}{
			[]string{"-f", "dates.ldg", "register"},
			"refdata/dates.register1.ref",
		},
		[]interface{}{
			[]string{"-f", "dates.ldg", "register", "Expenses"},
			"refdata/dates.register2.ref",
		},
		[]interface{}{
			[]string{"-f", "dates.ldg", "register", "Expenses:Sta"},
			"refdata/dates.register3.ref",
		},
		[]interface{}{
			[]string{"-f", "dates.ldg", "equity"},
			"refdata/dates.equity.ref",
		},
	}
	for _, testcase := range testcases {
		ref := testdataFile(testcase[1].(string))
		args := testcase[0].([]string)
		cmd := exec.Command(LEDGEREXEC, args...)
		out, _ := cmd.CombinedOutput()
		if updateref {
			ioutil.WriteFile(testcase[1].(string), out, 0660)
		}
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
			"refdata/date7.register.ref",
		},
		[]interface{}{
			[]string{"-f", "date7.ldg", "equity"},
			"refdata/date7.equity.ref",
		},
	}
	for _, testcase := range testcases {
		ref := testdataFile(testcase[1].(string))
		args := testcase[0].([]string)
		cmd := exec.Command(LEDGEREXEC, args...)
		out, _ := cmd.CombinedOutput()
		if updateref {
			ioutil.WriteFile(testcase[1].(string), out, 0660)
		}
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
	//		"refdata/drewr3.balance.ref",
	//	},
	//}
	//for _, testcase := range testcases {
	//	ref := testdataFile(testcase[1].(string))
	//	cmd := exec.Command(LEDGEREXEC, testcase[0].([]string)...)
	//	out, _ := cmd.CombinedOutput()
	//	fmt.Println(testcase[0], "................")
	//	fmt.Println(string(ref))
	//	if bytes.Compare(out, ref) != 0 {
	//		t.Logf(strings.Join(args, " "))
	//		t.Logf("expected %q", ref)
	//		t.Errorf("got %q", out)
	//	}
	//}
}

func TestFirst(t *testing.T) {
	testcases := [][]interface{}{
		[]interface{}{
			[]string{"-f", "first.ldg", "balance"},
			"refdata/first.balance1.ref",
		},
		[]interface{}{
			[]string{"-f", "first.ldg", "balance", "Assets"},
			"refdata/first.balance2.ref",
		},
		[]interface{}{
			[]string{"-f", "first.ldg", "balance", "Expenses"},
			"refdata/first.balance3.ref",
		},
		[]interface{}{
			[]string{"-f", "first.ldg", "register"},
			"refdata/first.register1.ref",
		},
		[]interface{}{
			[]string{"-f", "first.ldg", "register", "Expens|Check"},
			"refdata/first.register2.ref",
		},
		[]interface{}{
			[]string{"-f", "first.ldg", "register", "@", "KFC"},
			"refdata/first.register3.ref",
		},
		[]interface{}{
			[]string{"-f", "first.ldg", "register", "@", "^KFC"},
			"refdata/first.register4.ref",
		},
		[]interface{}{
			[]string{"-f", "first.ldg", "register", "Assets"},
			"refdata/first.register5.ref",
		},
		[]interface{}{
			[]string{"-f", "first.ldg", "register", "Check"},
			"refdata/first.register6.ref",
		},
		[]interface{}{
			[]string{"-f", "first.ldg", "register", "Dinning"},
			"refdata/first.register7.ref",
		},
		[]interface{}{
			[]string{"-f", "first.ldg", "register", "Expenses"},
			"refdata/first.register8.ref",
		},
		[]interface{}{
			[]string{"-f", "first.ldg", "register", "Expenses:Din"},
			"refdata/first.register9.ref",
		},
		[]interface{}{
			[]string{"-f", "first.ldg", "register", "Expenses:Sta"},
			"refdata/first.register10.ref",
		},
		[]interface{}{
			[]string{"-f", "first.ldg", "register", "Expens|Check"},
			"refdata/first.register11.ref",
		},
		[]interface{}{
			[]string{"-f", "first.ldg", "register", "^Check"},
			"refdata/first.register12.ref",
		},
		[]interface{}{
			[]string{"-f", "first.ldg", "register", "^nses"},
			"refdata/first.register13.ref",
		},
		[]interface{}{
			[]string{"-f", "first.ldg", "register", "nses"},
			"refdata/first.register14.ref",
		},
		[]interface{}{
			[]string{"-f", "first.ldg", "equity"},
			"refdata/first.equity.ref",
		},
	}
	for _, testcase := range testcases {
		ref := testdataFile(testcase[1].(string))
		args := testcase[0].([]string)
		cmd := exec.Command(LEDGEREXEC, args...)
		out, _ := cmd.CombinedOutput()
		if updateref {
			ioutil.WriteFile(testcase[1].(string), out, 0660)
		}
		if bytes.Compare(out, ref) != 0 {
			t.Logf(strings.Join(args, " "))
			t.Logf("expected %s", ref)
			t.Errorf("got %s", out)
		}
	}
}

func TestReimburse(t *testing.T) {
	testcases := [][]interface{}{
		[]interface{}{
			[]string{"-f", "reimburse.ldg", "balance"},
			"refdata/reimburse.balance1.ref",
		},
		[]interface{}{
			[]string{"-f", "reimburse.ldg", "-dc", "balance"},
			"refdata/reimburse.dcbalance.ref",
		},
		[]interface{}{
			[]string{"-f", "reimburse.ldg", "register"},
			"refdata/reimburse.register1.ref",
		},
		[]interface{}{
			[]string{"-f", "reimburse.ldg", "equity"},
			"refdata/reimburse.equity.ref",
		},
	}
	for _, testcase := range testcases {
		ref := testdataFile(testcase[1].(string))
		args := testcase[0].([]string)
		cmd := exec.Command(LEDGEREXEC, args...)
		out, _ := cmd.CombinedOutput()
		if updateref {
			ioutil.WriteFile(testcase[1].(string), out, 0660)
		}
		if bytes.Compare(out, ref) != 0 {
			t.Logf(strings.Join(args, " "))
			t.Logf("expected %s", ref)
			t.Errorf("got %s", out)
		}
	}
}

func TestSecond(t *testing.T) {
	testcases := [][]interface{}{
		[]interface{}{
			[]string{"-f", "second.ldg", "balance"},
			"refdata/second.balance1.ref",
		},
		[]interface{}{
			[]string{"-f", "second.ldg", "balance",
				"^Assets", "^Liabilities"},
			"refdata/second.balance2.ref",
		},
		[]interface{}{
			[]string{"-f", "second.ldg", "balance",
				"Assets", "Liabilities.*"},
			"refdata/second.balance3.ref",
		},
		[]interface{}{
			[]string{"-f", "second.ldg", "balance",
				"Assets", "Liabilities"},
			"refdata/second.balance4.ref",
		},
		[]interface{}{
			[]string{"-f", "second.ldg", "balance",
				"Assets", "Liabilities.*"},
			"refdata/second.balance5.ref",
		},
		[]interface{}{
			[]string{"-f", "second.ldg", "balance",
				"^Assets", "^Liabilities"},
			"refdata/second.balance6.ref",
		},
		[]interface{}{
			[]string{"-f", "second.ldg", "balance",
				"^assets", "^liabilities"},
			"refdata/second.balance7.ref",
		},
	}
	for _, testcase := range testcases {
		ref := testdataFile(testcase[1].(string))
		args := testcase[0].([]string)
		cmd := exec.Command(LEDGEREXEC, args...)
		out, _ := cmd.CombinedOutput()
		if updateref {
			ioutil.WriteFile(testcase[1].(string), out, 0660)
		}
		if bytes.Compare(out, ref) != 0 {
			t.Logf(strings.Join(args, " "))
			t.Logf("expected %s", ref)
			t.Errorf("got %s", out)
		}
	}
}

func TestMixedComm(t *testing.T) {
	testcases := [][]interface{}{
		[]interface{}{
			[]string{"-f", "mixedcomm1.ldg", "balance"},
			"refdata/mixedcomm1.balance.ref",
		},
		[]interface{}{
			[]string{"-f", "mixedcomm1.ldg", "-dc", "balance"},
			"refdata/mixedcomm1.dcbalance.ref",
		},
		[]interface{}{
			[]string{"-f", "mixedcomm1.ldg", "register"},
			"refdata/mixedcomm1.register.ref",
		},
		[]interface{}{
			[]string{"-f", "mixedcomm1.ldg", "equity"},
			"refdata/mixedcomm1.equity.ref",
		},
		[]interface{}{
			[]string{"-f", "mixedcomm2.ldg", "balance"},
			"refdata/mixedcomm2.balance.ref",
		},
		[]interface{}{
			[]string{"-f", "mixedcomm2.ldg", "register"},
			"refdata/mixedcomm2.register.ref",
		},
		[]interface{}{
			[]string{"-f", "mixedcomm2.ldg", "equity"},
			"refdata/mixedcomm2.equity.ref",
		},
		[]interface{}{
			[]string{"-f", "commname.ldg", "balance"},
			"refdata/commname.balance.ref",
		},
	}
	for _, testcase := range testcases {
		ref := testdataFile(testcase[1].(string))
		args := testcase[0].([]string)
		cmd := exec.Command(LEDGEREXEC, args...)
		out, _ := cmd.CombinedOutput()
		if updateref {
			ioutil.WriteFile(testcase[1].(string), out, 0660)
		}
		if bytes.Compare(out, ref) != 0 {
			t.Logf(strings.Join(args, " "))
			t.Logf("expected %s", ref)
			t.Errorf("got %s", out)
		}
	}
}

func TestUnbalanced(t *testing.T) {
	testcases := [][]interface{}{
		[]interface{}{
			[]string{"-f", "emptytrans.ldg", "balance"},
			"refdata/emptytrans.ref",
		},
		[]interface{}{
			[]string{"-f", "emptytrans.ldg", "register"},
			"refdata/emptytrans.ref",
		},
		[]interface{}{
			[]string{"-f", "unbalanced1.ldg", "balance"},
			"refdata/unbalanced1.ref",
		},
		[]interface{}{
			[]string{"-f", "unbalanced1.ldg", "register"},
			"refdata/unbalanced1.ref",
		},
		[]interface{}{
			[]string{"-f", "unbalanced2.ldg", "balance"},
			"refdata/unbalanced2.ref",
		},
		[]interface{}{
			[]string{"-f", "unbalanced2.ldg", "register"},
			"refdata/unbalanced2.ref",
		},
		[]interface{}{
			[]string{"-f", "unbalanced3.ldg", "balance"},
			"refdata/unbalanced3.ref",
		},
	}
	for _, testcase := range testcases {
		ref := testdataFile(testcase[1].(string))
		args := testcase[0].([]string)
		cmd := exec.Command(LEDGEREXEC, args...)
		out, _ := cmd.CombinedOutput()
		if updateref {
			ioutil.WriteFile(testcase[1].(string), out, 0660)
		}
		if bytes.Compare(out, ref) != 0 {
			t.Logf(strings.Join(args, " "))
			t.Logf("expected %s", ref)
			t.Errorf("got %s", out)
		}
	}
}

func TestTrip(t *testing.T) {
	testcases := [][]interface{}{
		[]interface{}{
			[]string{"-f", "trip.ldg", "balance"},
			"refdata/trip.balance.ref",
		},
		[]interface{}{
			[]string{"-f", "trip.ldg", "-dc", "balance"},
			"refdata/trip.dcbalance.ref",
		},
		[]interface{}{
			[]string{"-f", "trip.ldg", "register"},
			"refdata/trip.register.ref",
		},
		[]interface{}{
			[]string{"-f", "trip.ldg", "equity"},
			"refdata/trip.equity.ref",
		},
	}
	for _, testcase := range testcases {
		ref := testdataFile(testcase[1].(string))
		args := testcase[0].([]string)
		cmd := exec.Command(LEDGEREXEC, args...)
		out, _ := cmd.CombinedOutput()
		if updateref {
			ioutil.WriteFile(testcase[1].(string), out, 0660)
		}
		if bytes.Compare(out, ref) != 0 {
			t.Logf(strings.Join(args, " "))
			t.Logf("expected %s", ref)
			t.Errorf("got %s", out)
		}
	}
}

func TestPostingErr(t *testing.T) {
	testcases := [][]interface{}{
		[]interface{}{
			[]string{"-f", "postingerr.ldg", "balance"},
			"refdata/postingerr.ref",
		},
		[]interface{}{
			[]string{"-f", "postingerr.ldg", "register"},
			"refdata/postingerr.ref",
		},
	}
	for _, testcase := range testcases {
		ref := testdataFile(testcase[1].(string))
		args := testcase[0].([]string)
		cmd := exec.Command(LEDGEREXEC, args...)
		out, _ := cmd.CombinedOutput()
		if updateref {
			ioutil.WriteFile(testcase[1].(string), out, 0660)
		}
		if bytes.Compare(out, ref) != 0 {
			t.Logf(strings.Join(args, " "))
			t.Logf("expected %s", ref)
			t.Errorf("got %s", out)
		}
	}
}

func TestLotPrice(t *testing.T) {
	testcases := [][]interface{}{
		[]interface{}{
			[]string{"-f", "lotpriceerr.ldg", "balance"},
			"refdata/lotpriceerr.ref",
		},
		[]interface{}{
			[]string{"-f", "lotprice.ldg", "balance"},
			"refdata/lotprice.balance.ref",
		},
		[]interface{}{
			[]string{"-f", "lotprice.ldg", "register"},
			"refdata/lotprice.register.ref",
		},
		[]interface{}{
			[]string{"-f", "lotprice.ldg", "equity"},
			"refdata/lotprice.equity.ref",
		},
	}
	for _, testcase := range testcases {
		ref := testdataFile(testcase[1].(string))
		args := testcase[0].([]string)
		cmd := exec.Command(LEDGEREXEC, args...)
		out, _ := cmd.CombinedOutput()
		if updateref {
			ioutil.WriteFile(testcase[1].(string), out, 0660)
		}
		if bytes.Compare(out, ref) != 0 {
			t.Logf(strings.Join(args, " "))
			t.Logf("expected %s", ref)
			t.Errorf("got %s", out)
		}
	}
}

func TestAcctree(t *testing.T) {
	testcases := [][]interface{}{
		[]interface{}{
			[]string{"-f", "acctree.ldg", "balance"},
			"refdata/acctree.balance.ref",
		},
		[]interface{}{
			[]string{"-f", "acctree.ldg", "-dc", "balance"},
			"refdata/acctree.dcbalance.ref",
		},
		[]interface{}{
			[]string{"-f", "acctree.ldg", "-nosubtotal", "balance"},
			"refdata/acctree.balance.nosubtotal.ref",
		},
		[]interface{}{
			[]string{"-f", "acctree.ldg", "register"},
			"refdata/acctree.register.ref",
		},
		[]interface{}{
			[]string{"-f", "acctree.ldg", "equity"},
			"refdata/acctree.equity.ref",
		},
	}
	for _, testcase := range testcases {
		ref := testdataFile(testcase[1].(string))
		args := testcase[0].([]string)
		cmd := exec.Command(LEDGEREXEC, args...)
		out, _ := cmd.CombinedOutput()
		if updateref {
			ioutil.WriteFile(testcase[1].(string), out, 0660)
		}
		if bytes.Compare(out, ref) != 0 {
			t.Logf(strings.Join(args, " "))
			t.Logf("expected %s", ref)
			t.Errorf("got %s", out)
		}
	}
}

func TestShare(t *testing.T) {
	testcases := [][]interface{}{
		[]interface{}{
			[]string{"-f", "share.ldg", "balance"},
			"refdata/share.balance.ref",
		},
		[]interface{}{
			[]string{"-f", "share.ldg", "register"},
			"refdata/share.register.ref",
		},
		[]interface{}{
			[]string{"-f", "share.ldg", "equity"},
			"refdata/share.equity.ref",
		},
	}
	for _, testcase := range testcases {
		ref := testdataFile(testcase[1].(string))
		args := testcase[0].([]string)
		cmd := exec.Command(LEDGEREXEC, args...)
		out, _ := cmd.CombinedOutput()
		if updateref {
			ioutil.WriteFile(testcase[1].(string), out, 0660)
		}
		if bytes.Compare(out, ref) != 0 {
			t.Logf(strings.Join(args, " "))
			t.Logf("expected %s", ref)
			t.Errorf("got %s", out)
		}
	}
}

func TestDirtAccount(t *testing.T) {
	testcases := [][]interface{}{
		[]interface{}{
			[]string{"-f", "dirtaccount1.ldg", "-strict", "list", "accounts"},
			"refdata/dirtaccount1.list.ref",
		},
		[]interface{}{
			[]string{"-f", "dirtaccount1.ldg", "-strict", "-v", "list",
				"accounts"},
			"refdata/dirtaccount1.vlist.ref",
		},
		[]interface{}{
			[]string{"-f", "dirtaccount2.ldg", "-strict", "list", "accounts"},
			"refdata/dirtaccount2.list.ref",
		},
		[]interface{}{
			[]string{"-f", "dirtaccount2.ldg", "-strict", "-v", "list",
				"accounts"},
			"refdata/dirtaccount2.vlist.ref",
		},
		[]interface{}{
			[]string{"-f", "dirtwithacc.ldg", "-strict", "balance"},
			"refdata/dirtwithacc.balance.ref",
		},
		[]interface{}{
			[]string{"-f", "dirtwithacc.ldg", "-strict", "register"},
			"refdata/dirtwithacc.register.ref",
		},
		[]interface{}{
			[]string{"-f", "dirtwithoacc.ldg", "-strict", "balance"},
			"refdata/dirtwithoacc.balance.ref",
		},
		[]interface{}{
			[]string{"-f", "dirtwithoacc.ldg", "-strict", "register"},
			"refdata/dirtwithoacc.register.ref",
		},
	}
	for _, testcase := range testcases {
		ref := testdataFile(testcase[1].(string))
		args := testcase[0].([]string)
		cmd := exec.Command(LEDGEREXEC, args...)
		out, _ := cmd.CombinedOutput()
		if updateref {
			ioutil.WriteFile(testcase[1].(string), out, 0660)
		}
		if bytes.Compare(out, ref) != 0 {
			t.Logf(strings.Join(args, " "))
			t.Logf("expected %s", ref)
			t.Errorf("got %s", out)
		}
	}
}

func TestMatchingPayee(t *testing.T) {
	testcases := [][]interface{}{
		[]interface{}{
			[]string{"-f", "unmatchpayee.ldg", "balance"},
			"refdata/unmatchpayee.ref",
		},
		[]interface{}{
			[]string{"-f", "nomatchpayee.ldg", "balance"},
			"refdata/nomatchpayee.ref",
		},
		[]interface{}{
			[]string{"-f", "matchpayee.ldg", "balance"},
			"refdata/matchpayee.balance.ref",
		},
		[]interface{}{
			[]string{"-f", "matchpayee.ldg", "register"},
			"refdata/matchpayee.register.ref",
		},
	}
	for _, testcase := range testcases {
		ref := testdataFile(testcase[1].(string))
		args := testcase[0].([]string)
		cmd := exec.Command(LEDGEREXEC, args...)
		out, _ := cmd.CombinedOutput()
		if updateref {
			ioutil.WriteFile(testcase[1].(string), out, 0660)
		}
		if bytes.Compare(out, ref) != 0 {
			t.Logf(strings.Join(args, " "))
			t.Logf("expected %s", ref)
			t.Errorf("got %s", out)
		}
	}
}

func TestDirectiveBucket(t *testing.T) {
	testcases := [][]interface{}{
		[]interface{}{
			[]string{"-f", "dirtbucket.ldg", "balance"},
			"refdata/dirtbucket.balance.ref",
		},
		[]interface{}{
			[]string{"-f", "dirtbucket.ldg", "register"},
			"refdata/dirtbucket.register.ref",
		},
	}
	for _, testcase := range testcases {
		ref := testdataFile(testcase[1].(string))
		args := testcase[0].([]string)
		cmd := exec.Command(LEDGEREXEC, args...)
		out, _ := cmd.CombinedOutput()
		if updateref {
			ioutil.WriteFile(testcase[1].(string), out, 0660)
		}
		if bytes.Compare(out, ref) != 0 {
			t.Logf(strings.Join(args, " "))
			t.Logf("expected %s", ref)
			t.Errorf("got %s", out)
		}
	}
}

func TestDirectiveCapture(t *testing.T) {
	testcases := [][]interface{}{
		[]interface{}{
			[]string{"-f", "dirtcapture.ldg", "balance"},
			"refdata/dirtcapture.balance.ref",
		},
		[]interface{}{
			[]string{"-f", "dirtcapture.ldg", "register"},
			"refdata/dirtcapture.register.ref",
		},
	}
	for _, testcase := range testcases {
		ref := testdataFile(testcase[1].(string))
		args := testcase[0].([]string)
		cmd := exec.Command(LEDGEREXEC, args...)
		out, _ := cmd.CombinedOutput()
		if updateref {
			ioutil.WriteFile(testcase[1].(string), out, 0660)
		}
		if bytes.Compare(out, ref) != 0 {
			t.Logf(strings.Join(args, " "))
			t.Logf("expected %s", ref)
			t.Errorf("got %s", out)
		}
	}
}

func TestDirtCommodity(t *testing.T) {
	testcases := [][]interface{}{
		[]interface{}{
			[]string{"-f", "dirtcomm1.ldg", "-strict", "list", "commodity"},
			"refdata/dirtcomm1.list.ref",
		},
		[]interface{}{
			[]string{"-f", "dirtcomm1.ldg", "-strict", "-v", "list",
				"commodity"},
			"refdata/dirtcomm1.vlist.ref",
		},
		[]interface{}{
			[]string{"-f", "dirtcomm2.ldg", "-strict", "list", "commodity"},
			"refdata/dirtcomm2.list.ref",
		},
		[]interface{}{
			[]string{"-f", "dirtcomm2.ldg", "-strict", "-v", "list",
				"commodity"},
			"refdata/dirtcomm2.vlist.ref",
		},
		[]interface{}{
			[]string{"-f", "dirtcommerr1.ldg", "-strict", "list", "commodity"},
			"refdata/dirtcommerr1.ref",
		},
		[]interface{}{
			[]string{"-f", "dirtwithcomm.ldg", "-strict", "balance"},
			"refdata/dirtwithcomm.balance.ref",
		},
		[]interface{}{
			[]string{"-f", "dirtwithcomm.ldg", "-strict", "register"},
			"refdata/dirtwithcomm.register.ref",
		},
		[]interface{}{
			[]string{"-f", "dirtwithocomm.ldg", "-strict", "balance"},
			"refdata/dirtwithocomm.balance.ref",
		},
		[]interface{}{
			[]string{"-f", "dirtwithocomm.ldg", "-strict", "register"},
			"refdata/dirtwithocomm.register.ref",
		},
	}
	for _, testcase := range testcases {
		ref := testdataFile(testcase[1].(string))
		args := testcase[0].([]string)
		cmd := exec.Command(LEDGEREXEC, args...)
		out, _ := cmd.CombinedOutput()
		if updateref {
			ioutil.WriteFile(testcase[1].(string), out, 0660)
		}
		if bytes.Compare(out, ref) != 0 {
			t.Logf(strings.Join(args, " "))
			t.Logf("expected %s", ref)
			t.Errorf("got %s", out)
		}
	}
}

func TestDirtAlias(t *testing.T) {
	testcases := [][]interface{}{
		[]interface{}{
			[]string{"-f", "dirtalias.ldg", "-strict", "list", "accounts"},
			"refdata/dirtalias.list.ref",
		},
		[]interface{}{
			[]string{"-f", "dirtalias.ldg", "-strict", "-v", "list", "accounts"},
			"refdata/dirtalias.vlist.ref",
		},
		[]interface{}{
			[]string{"-f", "dirtaliaserr.ldg", "-strict", "list", "accounts"},
			"refdata/dirtaliaserr.ref",
		},
	}
	for _, testcase := range testcases {
		ref := testdataFile(testcase[1].(string))
		args := testcase[0].([]string)
		cmd := exec.Command(LEDGEREXEC, args...)
		out, _ := cmd.CombinedOutput()
		if updateref {
			ioutil.WriteFile(testcase[1].(string), out, 0660)
		}
		if bytes.Compare(out, ref) != 0 {
			t.Logf(strings.Join(args, " "))
			t.Logf("expected %s", ref)
			t.Errorf("got %s", out)
		}
	}
}

func TestDirtInclude(t *testing.T) {
	testcases := [][]interface{}{
		[]interface{}{
			[]string{"-f", "including.ldg", "balance"},
			"refdata/including.balance.ref",
		},
		[]interface{}{
			[]string{"-f", "including.ldg", "register"},
			"refdata/including.register.ref",
		},
	}
	for _, testcase := range testcases {
		ref := testdataFile(testcase[1].(string))
		args := testcase[0].([]string)
		cmd := exec.Command(LEDGEREXEC, args...)
		out, _ := cmd.CombinedOutput()
		if updateref {
			ioutil.WriteFile(testcase[1].(string), out, 0660)
		}
		if bytes.Compare(out, ref) != 0 {
			t.Logf(strings.Join(args, " "))
			t.Logf("expected %s", ref)
			t.Errorf("got %s", out)
		}
	}
}

// TODO: enable this once all command line options are implemented.
//func TestHelp(t *testing.T) {
//	testcases := [][]interface{}{
//		[]interface{}{
//			[]string{"-h"},
//			"refdata/help.ref",
//		},
//	}
//	for _, testcase := range testcases {
//		ref := testdataFile(testcase[1].(string))
//		args := testcase[0].([]string)
//		cmd := exec.Command(LEDGEREXEC, args...)
//		out, _ := cmd.CombinedOutput()
//		//ioutil.WriteFile(testcase[1].(string), out, 0660)
//		if bytes.Compare(out, ref) != 0 {
//			t.Logf(strings.Join(args, " "))
//			t.Logf("expected %s", ref)
//			t.Errorf("got %s", out)
//		}
//	}
//}

func TestVersion(t *testing.T) {
	testcases := [][]interface{}{
		[]interface{}{
			[]string{"version"},
			"refdata/version.ref",
		},
	}
	for _, testcase := range testcases {
		ref := testdataFile(testcase[1].(string))
		args := testcase[0].([]string)
		cmd := exec.Command(LEDGEREXEC, args...)
		out, _ := cmd.CombinedOutput()
		if updateref {
			ioutil.WriteFile(testcase[1].(string), out, 0660)
		}
		if bytes.Compare(out, ref) != 0 {
			t.Logf(strings.Join(args, " "))
			t.Logf("expected %s", ref)
			t.Errorf("got %s", out)
		}
	}
}

func TestOutfile(t *testing.T) {
	tempdir, ofile := os.TempDir(), "output"
	outfile := filepath.Join(tempdir, ofile)
	testcases := [][]interface{}{
		[]interface{}{
			[]string{"-f", "basic.ldg", "-o", outfile, "balance"},
			"refdata/basic.balance.ref",
		},
		[]interface{}{
			[]string{"-f", "basic.ldg", "-o", outfile, "register"},
			"refdata/basic.register.ref",
		},
		[]interface{}{
			[]string{"-f", "basic.ldg", "-o", outfile, "equity"},
			"refdata/basic.equity.ref",
		},
		[]interface{}{
			[]string{"-f", "dirtaccount1.ldg", "-strict", "-o", outfile,
				"list", "accounts"},
			"refdata/dirtaccount1.list.ref",
		},
		[]interface{}{
			[]string{"-f", "dirtaccount1.ldg", "-strict", "-v", "-o", outfile,
				"list", "accounts"},
			"refdata/dirtaccount1.vlist.ref",
		},
		[]interface{}{
			[]string{"-f", "dirtcomm1.ldg", "-strict", "-o", outfile,
				"list", "commodity"},
			"refdata/dirtcomm1.outlist.ref",
		},
		[]interface{}{
			[]string{"-f", "dirtcomm1.ldg", "-strict", "-v", "-o", outfile,
				"list", "commodity"},
			"refdata/dirtcomm1.outvlist.ref",
		},
	}
	for _, testcase := range testcases {
		ref := testdataFile(testcase[1].(string))
		args := testcase[0].([]string)
		cmd := exec.Command(LEDGEREXEC, args...)
		cmd.CombinedOutput()
		data, err := ioutil.ReadFile(outfile)
		if err != nil {
			t.Error(err)
		}
		if updateref {
			ioutil.WriteFile(testcase[1].(string), data, 0660)
		}
		if bytes.Compare(data, ref) != 0 {
			t.Logf(strings.Join(args, " "))
			t.Logf("expected %s", ref)
			t.Errorf("got %s", data)
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
