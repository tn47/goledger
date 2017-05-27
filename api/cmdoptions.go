package api

import "os"

var Options struct {
	Dbname     string
	Journals   []string
	Currentdt  string
	Begindt    string
	Enddt      string
	Finyear    int
	Period     string
	Cleared    bool
	Uncleared  bool
	Pending    bool
	Onlyreal   bool
	Onlyactual bool
	Related    bool
	Dcformat   bool
	Strict     bool
	Pedantic   bool
	Checkpayee bool
	Verbose    bool
	Outfd      *os.File
	Loglevel   string
}
