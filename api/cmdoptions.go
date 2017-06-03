package api

import "os"
import "time"

var Options struct {
	Dbname     string
	Journals   []string
	Currentdt  string
	Begindt    *time.Time
	Enddt      *time.Time
	Finyear    int
	Period     string
	Nosubtotal bool
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
