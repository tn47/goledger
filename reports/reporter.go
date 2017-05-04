package reports

import "fmt"

import "github.com/prataprc/goledger/api"

func NewReporter(args []string) (reporter api.Reporter) {
	if len(args[0]) == 0 {
		return &DummyReporter{}
	}

	switch args[0] {
	case "balance":
		reporter = NewReportBalance(args)
	case "register":
		reporter = NewReportRegister(args)
	}
	return reporter
}

type DummyReporter struct{}

func (report *DummyReporter) Transaction(
	db api.Datastorer, trans api.Transactor) {

	return
}

func (report *DummyReporter) Posting(
	db api.Datastorer, trans api.Transactor,
	p api.Poster, account api.Accounter) {

	return
}

func (report *DummyReporter) BubblePosting(
	db api.Datastorer, trans api.Transactor,
	p api.Poster, account api.Accounter) {

	return
}

func (report *DummyReporter) Render(args []string) {
	return
}

func (report *DummyReporter) String() string {
	return fmt.Sprintf("DummyReporter")
}
