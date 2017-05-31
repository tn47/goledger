package dblentry

import "github.com/prataprc/goparsec"

var ytokHardSpace = parsec.TokenExact(` {2}|\t`, "HARDSPACE")
var ytokEqual = parsec.Atom("=", "EQUAL")
var ytokCurrency = parsec.Token(`[^0-9 \t\r\n.,;:?!/@+*/^&|=<>(){}\[\]-]+`, "CURRENCY")
var ytokAmount = parsec.Token(`[0-9,.-]+`, "AMOUNT")
var ytokCommodity = parsec.Token(`[^0-9 \t\r\n.,;:?!/@+*/^&|=<>(){}\[\]-]+`, "COMMODITY")
var ytokAssert = parsec.Atom("assert", "ASSERT")
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

var ytokPostacc1 = parsec.Token(`[a-zA-Z]([0-9a-zA-Z`+accchars+`]+ )*([  ]|[\t])`, "POSTACCN1")
var ytokPostnote = parsec.Token(";[^;]+", "POSTNOTE")
var ytokAt = parsec.Atom("@", "COSTAT")
var ytokAtat = parsec.Atom("@@", "COSTATAT")
var ytokOpenparan = parsec.Atom("{", "OPENPARAN")
var ytokOpenOpenparan = parsec.Atom("{{", "OPENOPENPARAN")
var ytokCloseparan = parsec.Atom("}", "CLOSECLOSEPARAN")
var ytokCloseCloseparan = parsec.Atom("}}", "CLOSEPARAN")
var ytokOpensqrt = parsec.Atom(`[`, "OPENSQRT")
var ytokClosesqrt = parsec.Atom(`]`, "CLOSESQRT")

//---- Price tokens

var ytokPrice = parsec.Token(";[^;]+", "TRANSPRICE")

//---- Directives
var ytokAccount = parsec.Atom("account", "DRTV_ACCOUNT")
var ytokAlias = parsec.Atom("alias", "DRTV_ACCOUNT_ALIAS")
var ytokPayee = parsec.Atom("payee", "DRTV_ACCOUNT_PAYEE")
var ytokCheck = parsec.Atom("check", "DRTV_ACCOUNT_CHECK")
var ytokEval = parsec.Atom("eval", "DRTV_ACCOUNT_EVAL")
var ytokType = parsec.Atom("type", "DRTV_ACCOUNT_TYPE")
var ytokValue = parsec.Token(".*", "DRTV_VALUE")

var ytokApply = parsec.Atom("apply", "DRTV_APPLY")
var ytokAliasname = parsec.Token("[^=]+", "DRTV_ALIASNAME")
var ytokBucket = parsec.Atom("bucket", "DRTV_BUCKET")
var ytokCapture = parsec.Atom("capture", "DRTV_CAPTURE")
var ytokComment = parsec.Atom("comment", "DRTV_COMMENT")
var ytokDirtCommodity = parsec.Atom("commodity", "DRTV_COMMODITY")
var ytokFormat = parsec.Atom("format", "DRTV_COMMODITY_FORMAT")
var ytokNomarket = parsec.Atom("nomarket", "DRTV_COMMODITY_NOMARKET")
var ytokDirtDefine = parsec.Atom("define", "DRTV_DEFINE")
var ytokDirtFixed = parsec.Atom("fixed", "DRTV_FIXED")
var ytokDirtInclude = parsec.Atom("include", "DRTV_FIXED")
var ytokPayeeAlias = parsec.Atom("alias", "DRTV_PAYEE_ALIAS")
var ytokPayeeUuid = parsec.Atom("uuid", "DRTV_PAYEE_UUID")
var ytokDirtTest = parsec.Atom("test", "DRTV_TEST")
var ytokEnd = parsec.Atom("end", "DRTV_END")
var ytokYear = parsec.Atom("year", "DRTV_YEAR")

var ytokNote = parsec.Atom("note", "DRTV_NOTE")
var ytokDefault = parsec.Atom("default", "DRTV_DEFAULT")

// tags
var ytokColon = parsec.Atom(":", "COLON")
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
