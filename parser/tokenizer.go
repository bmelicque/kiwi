package parser

import (
	"bufio"
	"io"
	"regexp"
)

type Position struct {
	Line int
	Col  int
}

type Loc struct {
	Start Position
	End   Position
}

type TokenKind int

const (
	Illegal TokenKind = iota
	EOF
	EOL

	Name
	PlaceholderToken // _
	NumberLiteral
	BooleanLiteral
	StringLiteral

	StringKeyword   // string
	NumberKeyword   // number
	BooleanKeyword  // boolean
	IfKeyword       // if
	ElseKeyword     //else
	MatchKeyword    // match
	CaseKeyword     // case
	ForKeyword      // for
	InKeyword       // in
	BreakKeyword    // break
	ContinueKeyword // continue
	ReturnKeyword   // return
	TryKeyword      // try
	ThrowKeyword    // throw
	CatchKeyword    // catch
	AsyncKeyword    // async
	AwaitKeyword    // await

	Add        // +
	Concat     // ++
	Sub        // -
	Mul        // *
	Pow        // **
	Div        // /
	Mod        // %
	LogicalAnd // &&
	LogicalOr  // ||
	Bang       // !
	BinaryAnd  // &
	BinaryOr   // |

	QuestionMark // ?

	Less         // <
	Greater      // >
	LessEqual    // <=
	GreaterEqual // >=
	Equal        // ==
	NotEqual     // !=

	Define         // ::
	Declare        // :=
	Assign         // =
	ExclusiveRange // ..
	InclusiveRange // ..=
	SlimArrow      // ->
	FatArrow       // =>

	AddAssign        // +=
	ConcatAssign     // ++=
	SubAssign        // -=
	MulAssign        // *=
	PowAssign        // **=
	DivAssign        // /=
	ModAssign        // %=
	LogicalAndAssign // &&=
	LogicalOrAssign  // ||=

	LeftBracket      // [
	RightBracket     // ]
	LeftParenthesis  // (
	RightParenthesis // )
	LeftBrace        // {
	RightBrace       // }

	Comma // ,
	Colon // :
	Dot   // .
)

type Token interface {
	Kind() TokenKind
	Text() string
	Loc() Loc
}

type token struct {
	kind TokenKind
	loc  Loc
}

func (t token) Kind() TokenKind {
	return t.kind
}

func (t token) Text() string {
	switch t.kind {
	case EOL:
		return "\n"
	case Add,
		Concat:
		return "+"
	case Sub:
		return "-"
	case Mul:
		return "*"
	case Pow:
		return "**"
	case Div:
		return "/"
	case Mod:
		return "%"
	case LogicalAnd:
		return "&&"
	case LogicalOr:
		return "||"
	case Less:
		return "<"
	case Greater:
		return ">"
	case LessEqual:
		return "<="
	case GreaterEqual:
		return ">="
	case Equal:
		return "==="
	case NotEqual:
		return "!=="
	default:
		return ""
	}
}

func (t token) Loc() Loc {
	return t.loc
}

type literal struct {
	kind  TokenKind
	value string
	loc   Loc
}

func (l literal) Kind() TokenKind {
	return l.kind
}

func (l literal) Text() string {
	return l.value
}

func (l literal) Loc() Loc {
	return l.loc
}

var blank = regexp.MustCompile(`^[\t\f\r ]+`)
var newLine = regexp.MustCompile(`^\s+`)
var number = regexp.MustCompile(`^\d+`)
var str = regexp.MustCompile(`^".+?[^\\]"`)
var doubleQuoteString = regexp.MustCompile(`^"(.*?)[^\\]"`)
var word = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9]*`)
var operator = regexp.MustCompile(`^(\+\+?|->?|\*\*?|/|%|::|:=|\.\.=?|=>|<=?|>=?|={1,2}|!=?|\|{1,2}|\?|&&?)`)
var punctuation = regexp.MustCompile(`^(\[|\]|,|:|\(|\)|\{|\}|_|\.)`)

func split(data []byte, atEOF bool) (advance int, token []byte, err error) {
	switch {
	case blank.Match(data):
		token = blank.Find(data)
	case newLine.Match(data):
		token = newLine.Find(data)
	case number.Match(data):
		token = number.Find(data)
	case str.Match(data):
		token = str.Find(data)
	case doubleQuoteString.Match(data):
		token = doubleQuoteString.Find(data)
	case word.Match(data):
		token = word.Find(data)
	case operator.Match(data):
		token = operator.Find(data)
	case punctuation.Match(data):
		token = punctuation.Find(data)
	}

	if len(token) != 0 {
		return len(token), token, nil
	}
	return 1, data[:1], nil
}

type tokenizer struct {
	scanner *bufio.Scanner
	cursor  Position
	token   Token
	ready   bool
}

type Tokenizer interface {
	Peek() Token
	Consume() Token
	DiscardLineBreaks()
}

func NewTokenizer(reader io.Reader) *tokenizer {
	scanner := bufio.NewScanner(reader)
	scanner.Split(split)
	return &tokenizer{scanner, Position{1, 1}, nil, false}
}

func (t *tokenizer) updateCursor(token string) {
	if list := regexp.MustCompile(`\n`).FindAllString(token, -1); list != nil {
		t.cursor.Line += len(list)
		t.cursor.Col = 1
		return
	}
	t.cursor.Col += len(token)
}

func makeToken(text string, loc Loc) Token {
	switch text {
	case "_":
		return literal{Name, text, loc}
	case "true", "false":
		return literal{BooleanLiteral, text, loc}
	case "string":
		return token{StringKeyword, loc}
	case "number":
		return token{NumberKeyword, loc}
	case "boolean":
		return token{BooleanKeyword, loc}
	case "if":
		return token{IfKeyword, loc}
	case "else":
		return token{ElseKeyword, loc}
	case "match":
		return token{MatchKeyword, loc}
	case "case":
		return token{CaseKeyword, loc}
	case "for":
		return token{ForKeyword, loc}
	case "in":
		return token{InKeyword, loc}
	case "break":
		return token{BreakKeyword, loc}
	case "continue":
		return token{ContinueKeyword, loc}
	case "return":
		return token{ReturnKeyword, loc}
	case "try":
		return token{TryKeyword, loc}
	case "throw":
		return token{ThrowKeyword, loc}
	case "catch":
		return token{CatchKeyword, loc}
	case "async":
		return token{AsyncKeyword, loc}
	case "await":
		return token{AwaitKeyword, loc}
	case "+":
		return token{Add, loc}
	case "++":
		return token{Concat, loc}
	case "-":
		return token{Sub, loc}
	case "*":
		return token{Mul, loc}
	case "**":
		return token{Pow, loc}
	case "/":
		return token{Div, loc}
	case "%":
		return token{Mod, loc}
	case "&&":
		return token{LogicalAnd, loc}
	case "||":
		return token{LogicalOr, loc}
	case "!":
		return token{Bang, loc}
	case "&":
		return token{BinaryAnd, loc}
	case "|":
		return token{BinaryOr, loc}
	case "<":
		return token{Less, loc}
	case ">":
		return token{Greater, loc}
	case "<=":
		return token{LessEqual, loc}
	case ">=":
		return token{GreaterEqual, loc}
	case "==":
		return token{Equal, loc}
	case "!=":
		return token{NotEqual, loc}
	case "?":
		return token{QuestionMark, loc}
	case "[":
		return token{LeftBracket, loc}
	case "]":
		return token{RightBracket, loc}
	case "(":
		return token{LeftParenthesis, loc}
	case ")":
		return token{RightParenthesis, loc}
	case "{":
		return token{LeftBrace, loc}
	case "}":
		return token{RightBrace, loc}
	case ",":
		return token{Comma, loc}
	case ":":
		return token{Colon, loc}
	case ".":
		return token{Dot, loc}
	case "::":
		return token{Define, loc}
	case ":=":
		return token{Declare, loc}
	case "=":
		return token{Assign, loc}
	case "..":
		return token{ExclusiveRange, loc}
	case "..=":
		return token{InclusiveRange, loc}
	case "->":
		return token{SlimArrow, loc}
	case "=>":
		return token{FatArrow, loc}
	}
	switch {
	case newLine.MatchString(text):
		return token{EOL, loc}
	case number.MatchString(text):
		return literal{NumberLiteral, text, loc}
	case str.MatchString(text):
		return literal{StringLiteral, text, loc}
	case word.MatchString(text):
		return literal{Name, text, loc}
	}
	return token{Illegal, loc}
}

func (t *tokenizer) next() bool {
	if t.token != nil {
		return true
	}
	if !t.scanner.Scan() {
		return false
	}
	value := t.scanner.Text()
	if blank.MatchString(value) {
		t.updateCursor(value)
		return t.next()
	}
	loc := Loc{t.cursor, Position{}}
	t.updateCursor(value)
	loc.End = t.cursor
	t.token = makeToken(value, loc)
	return true
}

func (t *tokenizer) Peek() Token {
	if t.next() {
		return t.token
	}
	return token{EOF, Loc{t.cursor, t.cursor}}
}

func (t *tokenizer) Consume() Token {
	if !t.next() {
		return token{EOF, Loc{t.cursor, t.cursor}}
	}
	token := t.token
	t.token = nil
	return token
}

func (t *tokenizer) DiscardLineBreaks() {
	token := t.Peek()
	for token.Kind() == EOL {
		t.Consume()
		token = t.Peek()
	}
}
