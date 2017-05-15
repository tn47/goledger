package api

import "fmt"

// Version type with {major-byte, minor-byte}.
type Version [2]byte

// LedgerVersion is current ledger version
var LedgerVersion = Version{0, 1}

func (ver Version) String() string {
	return fmt.Sprintf("%v.%v", ver[0], ver[1])
}
