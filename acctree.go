package main

import "github.com/tn47/goledger/dblentry"

type accnode struct {
	name     string
	children map[string]*accnode
}

func newaccnode(name string) *accnode {
	return &accnode{name: name, children: make(map[string]*accnode)}
}

func (an *accnode) adddescendants(names []string) {
	if len(names) == 0 {
		return
	}
	child, ok := an.children[names[0]]
	if ok == false {
		child = newaccnode(names[0])
	}
	an.children[names[0]] = child
	child.adddescendants(names[1:])
	return
}

func accpath2tree(accnames []string) *accnode {
	root := newaccnode("__root__")
	for _, accname := range accnames {
		root.adddescendants(dblentry.SplitAccount(accname))
	}
	return root
}
