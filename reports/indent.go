package reports

import "fmt"
import "strings"

import "github.com/tn47/goledger/api"

var _ = fmt.Sprintf("dummy")

func Indent(keys []string) []string {
	if len(keys) == 0 {
		return keys
	}
	// split each key
	kkeys := [][]string{}
	for _, key := range keys {
		kkeys = append(kkeys, strings.Split(key, ":"))
	}
	// Indent
	keys = []string{}
	for _, kkey := range doindent(kkeys[0], kkeys[1:]) {
		n := 0
		for _, k := range kkey {
			if k != "" {
				break
			}
			n++
		}
		s := api.Repeatstr("  ", n) + strings.Join(kkey[n:], ":")
		keys = append(keys, s)
	}
	return keys
}

func doindent(prefix []string, names [][]string) [][]string {
	outs := [][]string{prefix}

	//fmt.Println()
	//fmt.Printf("** doindent: %v ------ %v\n", prefix, names)
	if len(names) == 0 {
		//fmt.Printf("quick outs --- %v\n", outs)
		return outs
	}

	with, witho := partitionByPrefix(prefix, names)
	//fmt.Printf("** partitionByPrefix %v --- %v --- %v\n", prefix, with, witho)
	if len(with) > 0 {
		with1 := pruneprefix(with, prefix)
		//fmt.Printf("** prunedprefix %v --- %v\n", prefix, with1)
		outs = append(outs, prepend("", doindent(with1[0], with1[1:]))...)
		//fmt.Printf("** with-doindent %v --- %v\n", prefix, outs)
	}
	if len(witho) > 0 {
		outs = append(outs, doindent(witho[0], witho[1:])...)
		//fmt.Printf("** witho-doindent %v --- %v\n", prefix, outs)
	}
	//fmt.Printf("outs --- %v\n", outs)
	return outs
}

func pruneprefix(names [][]string, prefix []string) [][]string {
	pruned, x := [][]string{}, len(prefix)
	for _, name := range names {
		pruned = append(pruned, name[x:])
	}
	return pruned
}

func prepend(spaces string, names [][]string) [][]string {
	outs := [][]string{}
	for _, name := range names {
		outs = append(outs, append([]string{spaces}, name...))
	}
	return outs
}

func partitionByPrefix(
	prefix []string, names [][]string) ([][]string, [][]string) {

	with, witho := [][]string{}, [][]string{}

	for i, name := range names {
		for j, pref := range prefix {
			if pref != name[j] {
				witho = append(witho, names[i:]...)
				return with, witho
			}
		}
		with = append(with, name)
	}
	return with, witho // witho shall be ZERO len
}
