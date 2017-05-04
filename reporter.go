package main

import "github.com/prataprc/goledger/dblentry"

type Reporter interface {
	GetCallback() func(
		db *dblentry.Datastore, trans *dblentry.Transaction,
		p *dblentry.Posting, acc *dblentry.Account)

	Render(args []string)
}

func NewReporter(args []string) (reporter Reporter) {
	if len(args[0]) > 0 {
		return &DummyReporter{}
	}

	switch args[0] {
	case "balance":
		reporter = NewReportBalance(args)
	}
	return reporter
}

type DummyReporter struct{}

func (report *DummyReporter) GetCallback() func(
	db *dblentry.Datastore, trans *dblentry.Transaction,
	p *dblentry.Posting, acc *dblentry.Account) {

	return report.callback
}

func (report *DummyReporter) Render(args []string) {
	return
}

func (report *DummyReporter) callback(
	db *dblentry.Datastore, trans *dblentry.Transaction,
	p *dblentry.Posting, acc *dblentry.Account) {

	return
}
