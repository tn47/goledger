package dblentry

import "fmt"

import "github.com/prataprc/goparsec"

// Comment type
type Comment struct {
	line string
}

// NewComment create comment type.
func NewComment() *Comment {
	return &Comment{}
}

//---- ledger parser

// Yledger return a parser-combinator that can parse a comment string.
func (cmt *Comment) Yledger(db *Datastore) parsec.Parser {
	y := parsec.And(
		func(nodes []parsec.ParsecNode) parsec.ParsecNode {
			cmt.line = string(nodes[1].(*parsec.Terminal).Value)
			return cmt
		},
		ytokCommentchar, ytokCommentline,
	)
	return y
}

//---- engine

func (cmt *Comment) Firstpass(db *Datastore) error {
	return nil
}

func (cmt *Comment) Secondpass(db *Datastore) error {
	return fmt.Errorf("impossible situation")
}
