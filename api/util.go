package api

import "strings"
import "regexp"
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

// Filterstring can be used for excluding or including patters.
func Filterstring(strpatt string, strs []string) bool {
	if len(strs) == 0 {
		return true
	}
	for _, item := range strs {
		if strings.HasPrefix(strpatt, item) {
			return true
		}
		if ok, err := regexp.Match(item, []byte(strpatt)); err != nil {
			panic(err)
		} else if ok {
			return true
		}
	}
	return false
}
