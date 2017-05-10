package dblentry

import "github.com/prataprc/goparsec"

var ytok_equal = parsec.Token("=", "EQUAL")
var ytok_currency = parsec.Token(`[^0-9 \t\r\n.,;:?!/@+*/^&|=<>(){}\[\]-]+`, "CURRENCY")
var ytok_amount = parsec.Token(`[0-9,.-]+`, "AMOUNT")
var ytok_commodity = parsec.Token(`[^0-9 \t\r\n.,;:?!/@+*/^&|=<>(){}\[\]-]+`, "COMMODITY")
var ytok_assert = parsec.Token("assert", "ASSERT")
var ytok_expr = parsec.Token(`\{.*\}`, "EXPRESSION")
var ytok_yearval = parsec.Token(`[0-9]{4}`, "YEAR")
var ytok_commentchar = parsec.Token(`[;|#*]`, "COMMENTCHAR")
var ytok_commentline = parsec.Token(`.*`, "COMMENTLINE")

//---- Transaction tokens
type Transnote string

var accchars = `~!^()_\{}\[\]:;"'<>,.?/-`
var ytok_accname = parsec.Token(`[a-zA-Z][0-9a-zA-Z `+accchars+`]*`, "FULLACCNM")
var ytok_vaccname = parsec.Token(`\([a-zA-Z][0-9a-zA-Z: `+accchars+`]*\)`, "VFULLACCNM")
var ytok_baccname = parsec.Token(`\[[a-zA-Z][0-9a-zA-Z: `+accchars+`]*\]`, "BFULLACCNM")

var ytok_prefix = parsec.Token(`\*|!`, "TRANSPREFIX")
var ytok_code = parsec.Token(`\(.*\)`, "TRANSCODE")
var ytok_payeestr = parsec.Token("([^ \t]+[ ]?)+", "TRANSPAYEE")
var ytok_transnote = parsec.Token(";.*", "TRANSNOTE")

//---- Posting tokens

var ytok_postacc1 = parsec.Token(`[a-zA-Z]([0-9a-zA-Z`+accchars+`]* )*([  ]|[\t])`, "POSTACCN1")
var ytok_postnote = parsec.Token(";[^;]+", "POSTNOTE")
var ytok_at = parsec.Token("@", "COSTAT")
var ytok_atat = parsec.Token("@@", "COSTATAT")
var ytok_openparan = parsec.Token("{", "OPENPARAN")
var ytok_openopenparan = parsec.Token("{{", "OPENOPENPARAN")
var ytok_closeparan = parsec.Token("}", "CLOSECLOSEPARAN")
var ytok_closecloseparan = parsec.Token("}}", "CLOSEPARAN")
var ytok_opensqrt = parsec.Token(`\[`, "OPENSQRT")
var ytok_closesqrt = parsec.Token(`\]`, "CLOSESQRT")

//---- Price tokens

var ytok_price = parsec.Token(";[^;]+", "TRANSPRICE")

//---- Directives
var ytok_account = parsec.Token("account", "DRTV_ACCOUNT")
var ytok_note = parsec.Token("note", "DRTV_ACCOUNT_NOTE")
var ytok_alias = parsec.Token("alias", "DRTV_ACCOUNT_ALIAS")
var ytok_payee = parsec.Token("payee", "DRTV_ACCOUNT_PAYEE")
var ytok_check = parsec.Token("check", "DRTV_ACCOUNT_CHECK")
var ytok_eval = parsec.Token("eval", "DRTV_ACCOUNT_EVAL")
var ytok_default = parsec.Token("default", "DRTV_ACCOUNT_DEFAULT")
var ytok_value = parsec.Token(".*", "DRTV_VALUE")

var ytok_apply = parsec.Token("apply", "DRTV_APPLY")
var ytok_aliasname = parsec.Token("[^=]+", "DRTV_ALIASNAME")
var ytok_end = parsec.Token("end", "DRTV_END")
var ytok_year = parsec.Token("year", "DRTV_YEAR")

// tags
var ytok_colon = parsec.Token(":", "COLON")
var ytok_tag = parsec.Token(":[^ \t\r\n]", "TAG")
var ytok_tagk = parsec.Token("[^ \t\r\n]:[ \t]", "TAGKEY")
var ytok_tagv = parsec.Token(".+", "TAGVALUE")

//
func maybenode(nodes []parsec.ParsecNode) parsec.ParsecNode {
	if nodes == nil || len(nodes) == 0 {
		return nil
	}
	return nodes[0]
}

func Vector2scalar(nodes []parsec.ParsecNode) parsec.ParsecNode {
	return nodes[0]
}
