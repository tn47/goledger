package main

import "os"
import "fmt"
import "path"
import "bufio"
import "strings"
import "io/ioutil"
import "path/filepath"

import "github.com/prataprc/golog"

func readlines(filepath string) ([]string, error) {
	fd, err := os.Open(filepath)
	if err != nil {
		log.Errorf("%v\n", err)
		return nil, err
	}
	defer fd.Close()

	scanner := bufio.NewScanner(fd)
	scanner.Split(bufio.ScanLines)

	lines := []string{}
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, nil
}

func coveringjournals(cwd string) (files []string) {
	files = []string{}

	log.Debugf("gathering journals from %q\n", cwd)
	dirs := parentpaths(cwd, []string{})[1:]
	for _, dir := range dirs {
		entries, err := ioutil.ReadDir(dir)
		if err != nil {
			fmt.Printf("%v\n", err)
			continue
		}
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			filename := entry.Name()
			ok := filename == "ledgerrc"
			ok = ok || filename == ".ledgerrc"
			ok = ok || path.Ext(filename) == ".ldg"
			if ok {
				includefile := path.Join(dir, filename)
				log.Debugf("auto including %q\n", includefile)
				files = append(files, includefile)
			}
		}
	}
	return files
}

func listjournals(cwd string) ([]string, error) {
	files := []string{}
	items, err := ioutil.ReadDir(cwd)
	if err != nil {
		log.Errorf("%v\n", err)
		return nil, err
	}
	for _, item := range items {
		if path.Ext(item.Name()) == ".ldg" {
			files = append(files, filepath.Join(cwd, item.Name()))
		}
	}
	return files, nil
}

func findjournals(cwd string) ([]string, error) {
	files := []string{}
	filepath.Walk(
		cwd,
		func(pathdir string, info os.FileInfo, err error) error {
			if path.Ext(info.Name()) == ".ldg" {
				files = append(files, filepath.Join(pathdir, info.Name()))
			}
			return nil
		},
	)
	return files, nil
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
