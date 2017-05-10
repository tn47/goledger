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

func TestBasic(t *testing.T) {
	testcases := [][]interface{}{
		[]interface{}{
			[]string{"-f", "basic.ldg", "balance"},
			"refdata/basic.balance.ref",
		},
		[]interface{}{
			[]string{"-f", "basic.ldg", "register"},
			"refdata/basic.register.ref",
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
			[]string{"-f", "elidingamount2.ldg", "balance"},
			"refdata/elidingamount2.balance.ref",
		},
		[]interface{}{
			[]string{"-f", "elidingamount2.ldg", "register"},
			"refdata/elidingamount2.register.ref",
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
		//ioutil.WriteFile(testcase[1].(string), out, 0660)
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

func TestExplicitCost(t *testing.T) {
	testcases := [][]interface{}{
		[]interface{}{
			[]string{"-f", "explicitcost.ldg", "balance"},
			"refdata/explicitcost.balance.ref",
		},
		[]interface{}{
			[]string{"-f", "explicitcost.ldg", "register"},
			"refdata/explicitcost.register.ref",
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

func TestDates(t *testing.T) {
	testcases := [][]interface{}{
		[]interface{}{
			[]string{"-f", "dates.ldg", "balance"},
			"refdata/dates.balance.ref",
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
			"refdata/date7.register.ref",
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

func TestReimburse(t *testing.T) {
	testcases := [][]interface{}{
		[]interface{}{
			[]string{"-f", "reimburse.ldg", "balance"},
			"refdata/reimburse.balance1.ref",
		},
		[]interface{}{
			[]string{"-f", "reimburse.ldg", "register"},
			"refdata/reimburse.register1.ref",
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
		//ioutil.WriteFile(testcase[1].(string), out, 0660)
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
			[]string{"-f", "mixedcomm1.ldg", "register"},
			"refdata/mixedcomm1.register.ref",
		},
		[]interface{}{
			[]string{"-f", "mixedcomm2.ldg", "balance"},
			"refdata/mixedcomm2.balance.ref",
		},
		[]interface{}{
			[]string{"-f", "mixedcomm2.ldg", "register"},
			"refdata/mixedcomm2.register.ref",
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

func TestTrip(t *testing.T) {
	testcases := [][]interface{}{
		[]interface{}{
			[]string{"-f", "trip.ldg", "balance"},
			"refdata/trip.balance.ref",
		},
		[]interface{}{
			[]string{"-f", "trip.ldg", "register"},
			"refdata/trip.register.ref",
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
		//ioutil.WriteFile(testcase[1].(string), out, 0660)
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
