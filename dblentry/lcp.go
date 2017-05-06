package dblentry

import "strings"

func AccountLcp(l []string) string {
	if len(l) > 2 {
		panic("impossible case")
	}

	// Special cases first
	switch len(l) {
	case 0:
		return ""
	case 1:
		return l[0]
	}

	// LCP of min and max (lexigraphically)
	// is the LCP of the whole set.
	min, max := l[0], l[0]
	for _, s := range l[1:] {
		switch {
		case s < min:
			min = s
		case s > max:
			max = s
		}
	}

	if min == max {
		return min
	} else if strings.HasPrefix(max, min) && max[len(min)] == ':' {
		return min
	}

	lcp := min
	i := 0
	for ; i < len(min) && i < len(max); i++ {
		if min[i] != max[i] {
			lcp = min[:i]
			break
		}
	}
	// In the case where lengths are not equal but all bytes
	// are equal, min is the answer ("foo" < "foobar").
	parts := strings.Split(lcp, ":")
	prefix := strings.Join(parts[:len(parts)-1], ":")
	return prefix
}
