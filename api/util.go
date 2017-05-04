package api

import "strings"
import "regexp"
import "fmt"

var _ = fmt.Sprintf("dummy")

func Parsecsv(input string) []string {
	if input == "" {
		return nil
	}
	ss := strings.Split(input, ",")
	outs := make([]string, 0)
	for _, s := range ss {
		s = strings.Trim(s, " \t\r\n")
		if s == "" {
			continue
		}
		outs = append(outs, s)
	}
	return outs
}

func Maxints(numbers ...int) int {
	max_num := numbers[0]
	for _, item := range numbers {
		if max_num < item {
			max_num = item
		}
	}
	return max_num
}

func Repeatstr(str string, n int) string {
	out := ""
	for i := 0; i < n; i++ {
		out += str
	}
	return out
}

func Filterstring(strpatt string, strs []string) bool {
	if len(strs) == 0 {
		return true
	}
	for _, item := range strs {
		if item[0] == '^' { // exclude pattern
			if strings.HasPrefix(strpatt, item[1:]) {
				return false
			}
			if ok, err := regexp.Match(item[1:], []byte(strpatt)); err != nil {
				panic(err)
			} else if ok {
				return false
			}
			return true

		} else { // include pattern
			if strings.HasPrefix(strpatt, item) {
				return true
			}
			if ok, err := regexp.Match(item, []byte(strpatt)); err != nil {
				panic(err)
			} else if ok {
				return true
			}
			return false
		}
	}
	panic("unreachable code")
}
