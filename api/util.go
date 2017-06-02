package api

import "strings"
import "time"
import "fmt"

var _ = fmt.Sprintf("dummy")

// Parsecsv parses the input string for comma separated string values and
// return parsed strings.
func Parsecsv(input string) []string {
	if input == "" {
		return nil
	}
	ss := strings.Split(input, ",")

	var outs []string

	for _, s := range ss {
		s = strings.Trim(s, " \t\r\n")
		if s == "" {
			continue
		}
		outs = append(outs, s)
	}
	return outs
}

// Maxints return the max value amont numbers.
func Maxints(numbers ...int) int {
	maxNum := numbers[0]
	for _, item := range numbers {
		if maxNum < item {
			maxNum = item
		}
	}
	return maxNum
}

// Repeatstr to repeat the string `string` n times and return the same.
func Repeatstr(str string, n int) string {
	out := ""
	for i := 0; i < n; i++ {
		out += str
	}
	return out
}

func StringSet(xs []string) []string {
	// TODO: is there a better algorithm to identify duplicates
	ys := make([]string, len(xs))
outer:
	for _, x := range xs {
		for _, y := range ys {
			if x == y {
				continue outer
			}
		}
		ys = append(ys, x)
	}
	return ys
}

func ValidateDate(tm time.Time, year, month, date, hour, min, sec int) bool {
	y, m, d := tm.Date()
	h, t, s := tm.Clock()
	if y != year || m != time.Month(month) || d != date {
		return false
	} else if h != hour || t != min || s != sec {
		return false
	}
	return true
}
