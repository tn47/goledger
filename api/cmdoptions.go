package api

var Options struct {
	Dbname     string
	Journals   []string
	Currentdt  string
	Begindt    string
	Enddt      string
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
	Loglevel   string
}
