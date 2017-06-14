package dblentry

import "strings"

import "github.com/prataprc/goparsec"

// Tags from journal within a posting and/or a transactions.
type Tags struct {
	tags []string
	tagm map[string]interface{}
}

// NewTag create a new Tags instance.
func NewTag() *Tags {
	return &Tags{
		tags: []string{},
		tagm: make(map[string]interface{}),
	}
}

//---- ledger parser

// Yledger return a parser-combinator that can parse a tag specification.
func (tag *Tags) Yledger(db *Datastore) parsec.Parser {
	ytags := parsec.Kleene(
		func(nodes []parsec.ParsecNode) parsec.ParsecNode {
			if len(nodes) == 0 {
				return nil
			}
			tags := []string{}
			for _, node := range nodes {
				g := strings.Trim(string(node.(*parsec.Terminal).Value), ":")
				tags = append(tags, g)
			}
			return tags
		},
		ytokTag,
	)
	ytagscolon := parsec.And(
		func(nodes []parsec.ParsecNode) parsec.ParsecNode {
			return nodes[0].([]string)
		},
		ytags, ytokColon,
	)
	ytagm := parsec.And(
		func(nodes []parsec.ParsecNode) parsec.ParsecNode {
			key := strings.Trim(nodes[0].(*parsec.Terminal).Value, ": \t")
			val := strings.Trim(nodes[1].(*parsec.Terminal).Value, " \t")
			return map[string]interface{}{key: val}
		},
		ytokTagK, ytokTagV,
	)

	y := parsec.OrdChoice(
		func(nodes []parsec.ParsecNode) parsec.ParsecNode {
			switch obj := nodes[0].(type) {
			case []string:
				tag.tags = append(tag.tags, obj...)
			case map[string]interface{}:
				for k, v := range obj {
					tag.tagm[strings.ToLower(k)] = v
				}
			}
			return tag
		},
		ytagscolon, ytagm,
	)
	return y
}
