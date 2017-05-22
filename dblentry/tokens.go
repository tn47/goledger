package dblentry

import "github.com/prataprc/goparsec"

var ytokHardSpace = parsec.TokenStrict(` {2}|\t`, "HARDSPACE")
var ytokEqual = parsec.Token("=", "EQUAL")
var ytokCurrency = parsec.Token(`[^0-9 \t\r\n.,;:?!/@+*/^&|=<>(){}\[\]-]+`, "CURRENCY")
var ytokAmount = parsec.Token(`[0-9,.-]+`, "AMOUNT")
var ytokCommodity = parsec.Token(`[^0-9 \t\r\n.,;:?!/@+*/^&|=<>(){}\[\]-]+`, "COMMODITY")
var ytokAssert = parsec.Token("assert", "ASSERT")
var ytokExpr = parsec.Token(`\{.*\}`, "EXPRESSION")
var ytokYearval = parsec.Token(`[0-9]{4}`, "YEAR")
var ytokCommentchar = parsec.Token(`[;|#*]`, "COMMENTCHAR")
var ytokCommentline = parsec.Token(`.*`, "COMMENTLINE")

//---- Transaction tokens
type typeTransnote string

var accchars = `\&~!^()_\{}\[\]:;"'<>,.?/-`
var ytokAccname = parsec.Token(`[a-zA-Z][0-9a-zA-Z `+accchars+`]*`, "FULLACCNM")
var ytokVaccname = parsec.Token(`\([a-zA-Z][0-9a-zA-Z: `+accchars+`]*\)`, "VFULLACCNM")
var ytokBaccname = parsec.Token(`\[[a-zA-Z][0-9a-zA-Z: `+accchars+`]*\]`, "BFULLACCNM")

var ytokPrefix = parsec.Token(`\*|!`, "TRANSPREFIX")
var ytokCode = parsec.Token(`\(.*\)`, "TRANSCODE")
var ytokPayeestr = parsec.Token("([^ \t]+[ ]?)+", "TRANSPAYEE")
var ytokTransnote = parsec.Token(";.*", "TRANSNOTE")

//---- Posting tokens

var ytokPostacc1 = parsec.Token(`[a-zA-Z]([0-9a-zA-Z`+accchars+`]* )*([  ]|[\t])`, "POSTACCN1")
var ytokPostnote = parsec.Token(";[^;]+", "POSTNOTE")
var ytokAt = parsec.Token("@", "COSTAT")
var ytokAtat = parsec.Token("@@", "COSTATAT")
var ytokOpenparan = parsec.Token("{", "OPENPARAN")
var ytokOpenOpenparan = parsec.Token("{{", "OPENOPENPARAN")
var ytokCloseparan = parsec.Token("}", "CLOSECLOSEPARAN")
var ytokCloseCloseparan = parsec.Token("}}", "CLOSEPARAN")
var ytokOpensqrt = parsec.Token(`\[`, "OPENSQRT")
var ytokClosesqrt = parsec.Token(`\]`, "CLOSESQRT")

//---- Price tokens

var ytokPrice = parsec.Token(";[^;]+", "TRANSPRICE")

//---- Directives
var ytokAccount = parsec.Token("account", "DRTV_ACCOUNT")
var ytokAlias = parsec.Token("alias", "DRTV_ACCOUNT_ALIAS")
var ytokPayee = parsec.Token("payee", "DRTV_ACCOUNT_PAYEE")
var ytokCheck = parsec.Token("check", "DRTV_ACCOUNT_CHECK")
var ytokEval = parsec.Token("eval", "DRTV_ACCOUNT_EVAL")
var ytokValue = parsec.Token(".*", "DRTV_VALUE")

var ytokApply = parsec.Token("apply", "DRTV_APPLY")
var ytokAliasname = parsec.Token("[^=]+", "DRTV_ALIASNAME")
var ytokBucket = parsec.Token("bucket", "DRTV_BUCKET")
var ytokCapture = parsec.Token("capture", "DRTV_CAPTURE")
var ytokComment = parsec.Token("comment", "DRTV_COMMENT")
var ytokDirtCommodity = parsec.Token("commodity", "DRTV_COMMODITY")
var ytokFormat = parsec.Token("format", "DRTV_COMMODITY_FORMAT")
var ytokNomarket = parsec.Token("nomarket", "DRTV_COMMODITY_NOMARKET")
var ytokDirtDefine = parsec.Token("define", "DRTV_DEFINE")
var ytokDirtFixed = parsec.Token("fixed", "DRTV_FIXED")
var ytokDirtTest = parsec.Token("test", "DRTV_TEST")
var ytokEnd = parsec.Token("end", "DRTV_END")
var ytokYear = parsec.Token("year", "DRTV_YEAR")

var ytokNote = parsec.Token("note", "DRTV_NOTE")
var ytokDefault = parsec.Token("default", "DRTV_DEFAULT")

// tags
var ytokColon = parsec.Token(":", "COLON")
var ytokTag = parsec.Token(":[^ \t\r\n]", "TAG")
var ytokTagK = parsec.Token("[^ \t\r\n]:[ \t]", "TAGKEY")
var ytokTagV = parsec.Token(".+", "TAGVALUE")

//
func maybenode(nodes []parsec.ParsecNode) parsec.ParsecNode {
	if nodes == nil || len(nodes) == 0 {
		return nil
	}
	return nodes[0]
}

// Vector2scalar convert vector of arity ONE to scalar.
func Vector2scalar(nodes []parsec.ParsecNode) parsec.ParsecNode {
	return nodes[0]
}
