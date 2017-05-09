package dblentry

import "strings"

import "github.com/prataprc/goparsec"

type Tags struct {
	tags []string
	tagm map[string]interface{}
}

func NewTag() *Tags {
	return &Tags{
		tags: []string{},
		tagm: make(map[string]interface{}),
	}
}

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
		ytok_tag,
	)
	ytagscolon := parsec.And(
		func(nodes []parsec.ParsecNode) parsec.ParsecNode {
			return nodes[0].([]string)
		},
		ytags, ytok_colon,
	)
	ytagm := parsec.And(
		func(nodes []parsec.ParsecNode) parsec.ParsecNode {
			key := string(nodes[0].(*parsec.Terminal).Value)
			val := string(nodes[1].(*parsec.Terminal).Value)
			return map[string]interface{}{key: val}
		},
		ytok_tagk, ytok_tagv,
	)

	y := parsec.OrdChoice(
		func(nodes []parsec.ParsecNode) parsec.ParsecNode {
			switch obj := nodes[0].(type) {
			case []string:
				tag.tags = append(tag.tags, obj...)
			case map[string]interface{}:
				for k, v := range obj {
					tag.tagm[k] = v
				}
			}
			return tag
		},
		ytagscolon, ytagm,
	)
	return y
}
