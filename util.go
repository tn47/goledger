package main

import "strings"

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

func maxints(numbers ...int) int {
	max_num := numbers[0]
	for _, item := range numbers {
		if max_num < item {
			max_num = item
		}
	}
	return max_num
}

func repeatstr(str string, n int) string {
	out := ""
	for i := 0; i < n; i++ {
		out += str
	}
	return out
}
