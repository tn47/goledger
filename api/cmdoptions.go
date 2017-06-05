package api

import "os"
import "fmt"
import "time"

var _ = fmt.Sprintf("dummy")

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

func FilterPeriod(date time.Time) bool {
	begin, end := Options.Begindt, Options.Enddt
	if begin != nil && date.Before(*begin) {
		return false
	} else if end != nil && date.Before(*end) {
		return true
	} else if end == nil {
		return true
	}
	return false
}
