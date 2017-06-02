[![Build Status](https://travis-ci.org/tn47/goledger.svg?branch=master)](https://travis-ci.org/tn47/goledger)
[![GoDoc](https://godoc.org/github.com/tn47/goledger?status.png)](https://godoc.org/github.com/tn47/goledger)

Inspired by [ledger-cli](http://ledger-cli), goledger is a re-write of command
line ledger in golang, with the stated goals.

* Keep the tool command line friendly, whether or not GUI / web
interface are available.
* Defaults to Locale en_IN, until the tool becomes smart enough
to handle locale specific details automatically.
* Targeted for personal, small and medium enterprises.
* Keep to the spirit of ledger-cli as much as possible.

Basic Concepts
==============

* ``Credit the giver debit the receiver``.
* ``Commodity``, based accounting, where currency is also a commodity.
* ``Account``, holding one or more commodities.
* ``Posting``, commodities can be posted to an account. Posting can be
debit posting or credit posting, based on whether the account is the
source account (credit the giver) or target account (debit the receiver)
* ``Transaction``, involving 2 or more postings, where the sum of credit
postings should balance out the sum of debit postings.

Once comfortable with above concepts, your can start book-keeping with ledger.

Transactions and Directives
===========================

Book-keeping starts with the journal file, typically ending with ``.ldg``. A
ledger file can contain directives (sometimes called as declarations) or
transaction entries.

There can be any number of directives and any number of transaction entries.

Transactions have the following format:

```text
2011/03/15   Whole Food Market
    Expenses:Groceries   75.00
    Assets:Checking      -75
```

**line1**: Date and Payee
**line2**: Accountname, Amount
**line3**: Accountname, Amount

* Each transaction is associated with a date, and it is always advised to
enter transaction with older date before and newer date after.
* Payee gives a name to the transaction, other than that no special
meaning is attached to it.
* Posting with positive amount are treated as debit transaction, and the
posting's account is called target account.
* Posting with negative amount are treated as credit transaction, and the
posting's account is called source account.

Ledger-likes
============

* [ledger rewrite in haskell](https://github.com/simonmichael/hledger)
* [ledger rewrite on JVM](https://github.com/sn127/tackler)
* [ledger rewrite in scala](https://github.com/hrj/abandon)
* [ledger-cli](https://github.com/ledger), it is worth spending time reading
upon the ledger specification.

How to contribute
=================

* Pick an issue, or create an new issue. Provide adequate documentation for
the issue.
* Assign the issue or get it assigned.
* Work on the code, once finished, raise a pull request.
* Goledger is written in [golang](https://golang.org/), hence expected to follow the
global guidelines for writing go programs.
* If the changeset is more than few lines, please generate a
[report card](https://goreportcard.com/report/github.com/tn47/goledger).
