package reports

//import "github.com/tn47/goledger/dblentry"
//
//type accnode struct {
//	name     string
//	children []*accnode
//}
//
//func newaccnode(name string) *accnode {
//	return &accnode{name: name, children: make([]*accnode, 0)}
//}
//
//func (an *accnode) adddescendants(names []string) {
//	if len(names) == 0 {
//		return
//	}
//	child, ok := an.children[names[0]]
//	if ok == false {
//		child = newaccnode(names[0])
//	}
//	return child.adddescendants(names[1:])
//}
//
//func accpath2tree(accnames []string) *accnode {
//	root := newaccnode("__root__")
//	for _, accname := range accnames {
//		root.adddescendants(dblentry.SplitAccount(accname))
//	}
//	return root
//}
