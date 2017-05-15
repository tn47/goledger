package dblentry

// AccountLcp return longest common prefix for account names.
func AccountLcp(ss []string) string {
	// Special cases first
	switch len(ss) {
	case 0:
		return ""
	case 1:
		return ss[0]
	}
	// LCP of min and max (lexigraphically)
	// is the LCP of the whole set.
	min, max := ss[0], ss[0]
	for _, s := range ss[1:] {
		switch {
		case s < min:
			min = s
		case s > max:
			max = s
		}
	}
	for i := 0; i < len(min) && i < len(max); i++ {
		if min[i] != max[i] {
			return min[:i]
		}
	}
	return min
}
