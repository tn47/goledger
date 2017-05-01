package main

import "os"
import "fmt"
import "path"
import "bufio"
import "strings"
import "io/ioutil"

import "github.com/prataprc/golog"

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
	log.Debugf("gathering journals from %q\n", cwd)
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
				includefile := path.Join(dir, filename)
				log.Debugf("auto including %q\n", includefile)
				files = append(files, includefile)
			}
		}
	}
	return files
}

func parentpaths(dirpath string, acc []string) (dirs []string) {
	dir, _ := path.Split(dirpath)
	if dir != "" {
		acc = append(acc, dirpath)
		dir = strings.TrimRight(dir, string(os.PathSeparator))
		return parentpaths(dir, acc)
	}
	return acc
}
