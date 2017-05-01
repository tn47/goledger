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
