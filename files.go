package main

import "os"
import "fmt"
import "path"
import "bufio"
import "strings"
import "io/ioutil"

func readlines(filepath string) []string {
	fd, _ := os.Open(filepath)
	defer fd.Close()

	scanner := bufio.NewScanner(fd)
	scanner.Split(bufio.ScanLines)

	lines := []string{}
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines
}

func getjournals(cwd string) (files []string) {
	dirs := parentpaths(cwd, []string{})
	for _, dir := range dirs {
		entries, err := ioutil.ReadDir(dir)
		if err != nil {
			fmt.Printf("%v\n", err)
			continue
		}
		files = []string{}
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			filename := entry.Name()
			ok := filename == "ledgerrc"
			ok = ok || filename == ".ledgerrc"
			ok = ok || strings.HasPrefix(filename, "ledger_")
			if ok {
				files = append(files, path.Join(dir, filename))
			}
		}
	}
	return files
}

func parentpaths(dirpath string, acc []string) (dirs []string) {
	dir, _ := path.Split(dirpath)
	if dir != "" {
		return parentpaths(dir, append(acc, dirpath))
	}
	return acc
}
