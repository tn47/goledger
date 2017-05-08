package dblentry

import "fmt"

import "github.com/prataprc/goparsec"

type Comment struct {
	line string
}

func NewComment() *Comment {
	return &Comment{}
}

func (cmt *Comment) Yledger(db *Datastore) parsec.Parser {
	y := parsec.And(
		func(nodes []parsec.ParsecNode) parsec.ParsecNode {
			cmt.line = string(nodes[1].(*parsec.Terminal).Value)
			return cmt
		},
		ytok_commentchar, ytok_commentline,
	)
	return y
}

func (cmt *Comment) Firstpass(db *Datastore) error {
	return nil
}

func (cmt *Comment) Secondpass(db *Datastore) error {
	return fmt.Errorf("impossible situation")
}
