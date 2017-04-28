package main

import "os"
import "bufio"

var options struct {
	journal string
}

func main() {
	options.journal = "examples/journal.dat"

	readlines(options.journal)
}

func readlines(path string) []string {
	fd, _ := os.Open(path)
	defer fd.Close()

	scanner := bufio.NewScanner(fd)
	scanner.Split(bufio.ScanLines)

	lines := []string{}
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines
}
