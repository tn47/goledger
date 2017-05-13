package api

import "fmt"

type Version [2]byte

var LedgerVersion = Version{0, 1}

func (ver Version) String() string {
	return fmt.Sprintf("%v.%v", ver[0], ver[1])
}
