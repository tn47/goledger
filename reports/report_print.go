package reports

import "fmt"

import "github.com/tn47/goledger/api"

// ReportPrint for balance reporting.
type ReportPrint struct {
	transs []api.Transactor
}

// NewReportList creates an instance for balance reporting
func NewReportPrint(args []string) *ReportPrint {
	report := &ReportPrint{transs: make([]api.Transactor, 0)}
	return report
}

//---- api.Reporter methods

func (report *ReportPrint) Firstpass(
	db api.Datastorer, trans api.Transactor, p api.Poster) error {

	return nil
}

func (report *ReportPrint) Transaction(
	_ api.Datastorer, trans api.Transactor) error {

	date := trans.Date()
	if dt := api.Options.Begindt; dt != nil && date.Before(*dt) {
		return nil
	} else if dt = api.Options.Enddt; dt != nil && date.Before(*dt) {
		report.transs = append(report.transs, trans)
	} else {
		report.transs = append(report.transs, trans)
	}
	return nil
}

func (report *ReportPrint) Posting(
	_ api.Datastorer, _ api.Transactor, _ api.Poster) error {

	return nil
}

func (report *ReportPrint) BubblePosting(
	_ api.Datastorer, _ api.Transactor, _ api.Poster, _ api.Accounter) error {

	return nil
}

func (report *ReportPrint) Render(args []string, ndb api.Datastorer) {
	outfd := api.Options.Outfd
	for _, trans := range report.transs {
		for _, line := range trans.Printlines() {
			fmt.Fprintln(outfd, line)
		}
		fmt.Fprintln(outfd)
	}
}

func (report *ReportPrint) Clone() api.Reporter {
	nreport := *report
	nreport.transs = make([]api.Transactor, 0, len(report.transs))
	for _, trans := range report.transs {
		nreport.transs = append(nreport.transs, trans)
	}
	return &nreport
}

func (report *ReportPrint) Startjournal(fname string, included bool) {
	panic("not implemented")
}
