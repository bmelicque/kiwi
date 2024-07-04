package tokenizer

import (
	"bufio"
	"log"
	"os"
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
	ILLEGAL TokenKind = iota
	EOF
	EOL

	IDENTIFIER
	PLACEHOLDER // _
	NUMBER
	BOOLEAN
	STRING

	STR_KW    // string
	NUM_KW    // number
	BOOL_KW   // boolean
	IF_KW     // if
	FOR_KW    // for
	RETURN_KW // return

	ADD    // +
	CONCAT // ++
	SUB    // -
	MUL    // *
	POW    // **
	DIV    // /
	MOD    // %
	LAND   // &&
	LOR    // ||

	LESS    // <
	GREATER // >
	LEQ     // <=
	GEQ     // >=
	EQ      // ==
	NEQ     // !=

	DEFINE          // ::
	DECLARE         // :=
	ASSIGN          // =
	RANGE_EXCLUSIVE // ..
	RANGE_INCLUSIVE // ..=
	SLIM_ARR        // ->
	FAT_ARR         // =>

	LBRACKET // [
	RBRACKET // ]
	LPAREN   // (
	RPAREN   // )
	LBRACE   // {
	RBRACE   // }

	COMMA // ,
	COLON // :
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
	case ADD,
		CONCAT:
		return "+"
	case SUB:
		return "-"
	case MUL:
		return "*"
	case POW:
		return "**"
	case DIV:
		return "/"
	case MOD:
		return "%"
	case LAND:
		return "&&"
	case LOR:
		return "||"
	case LESS:
		return "<"
	case GREATER:
		return ">"
	case LEQ:
		return "<="
	case GEQ:
		return ">="
	case EQ:
		return "==="
	case NEQ:
		return "!=="
	// case "::":
	// 	return token{DEFINE, loc}
	// case ":=":
	// 	return token{DECLARE, loc}
	// case "..":
	// 	return token{RANGE_INCLUSIVE, loc}
	// case "..=":
	// 	return token{RANGE_EXCLUSIVE, loc}
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

var blank = regexp.MustCompile(`^\s+`)
var newLine = regexp.MustCompile(`^\n`)
var number = regexp.MustCompile(`^\d+`)
var doubleQuoteString = regexp.MustCompile(`^"(.*?)[^\\]"`)
var word = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9]*`)
var operator = regexp.MustCompile(`^(\+\+?|->?|\*\*?|/|%|::|:=|\.\.=?|=>|={1,2}|!=)`)
var punctuation = regexp.MustCompile(`^(\[|\]|,|:|\(|\)|\{|\}|_)`)

func split(data []byte, atEOF bool) (advance int, token []byte, err error) {
	switch {
	case blank.Match(data):
		token = blank.Find(data)
	case newLine.Match(data):
		token = newLine.Find(data)
	case number.Match(data):
		token = number.Find(data)
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
	file    *os.File
	scanner *bufio.Scanner
	cursor  Position
	token   Token
	ready   bool
}

type Tokenizer interface {
	Dispose()
	Peek() Token
	Consume() Token
	DiscardLineBreaks()
}

func New(path string) (*tokenizer, error) {
	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	scanner := bufio.NewScanner(bufio.NewReader(file))
	scanner.Split(split)
	return &tokenizer{file, scanner, Position{1, 1}, nil, false}, nil
}

func (t *tokenizer) Dispose() {
	t.file.Close()
}

func (t *tokenizer) updateCursor(token string) {
	// FIXME: line number is wrong... (count multiple '\n' once)
	if regexp.MustCompile(`\n`).MatchString(token) {
		t.cursor.Line++
		t.cursor.Col = 1
		return
	}
	t.cursor.Col += len(token)
}

func makeToken(text string, loc Loc) Token {
	switch text {
	case "\n":
		return token{EOL, loc}
	case "_":
		return literal{IDENTIFIER, text, loc}
	case "true", "false":
		return literal{BOOLEAN, text, loc}
	case "string":
		return token{STR_KW, loc}
	case "number":
		return token{NUM_KW, loc}
	case "boolean":
		return token{BOOL_KW, loc}
	case "if":
		return token{IF_KW, loc}
	case "for":
		return token{FOR_KW, loc}
	case "return":
		return token{RETURN_KW, loc}
	case "+":
		return token{ADD, loc}
	case "++":
		return token{CONCAT, loc}
	case "-":
		return token{SUB, loc}
	case "*":
		return token{MUL, loc}
	case "**":
		return token{POW, loc}
	case "/":
		return token{DIV, loc}
	case "%":
		return token{MOD, loc}
	case "&&":
		return token{LAND, loc}
	case "||":
		return token{LOR, loc}
	case "<":
		return token{LESS, loc}
	case ">":
		return token{GREATER, loc}
	case "<=":
		return token{LEQ, loc}
	case ">=":
		return token{GEQ, loc}
	case "==":
		return token{EQ, loc}
	case "!=":
		return token{NEQ, loc}
	case "[":
		return token{LBRACKET, loc}
	case "]":
		return token{RBRACKET, loc}
	case "(":
		return token{LPAREN, loc}
	case ")":
		return token{RPAREN, loc}
	case "{":
		return token{LBRACE, loc}
	case "}":
		return token{RBRACE, loc}
	case ",":
		return token{COMMA, loc}
	case ":":
		return token{COLON, loc}
	case "::":
		return token{DEFINE, loc}
	case ":=":
		return token{DECLARE, loc}
	case "=":
		return token{ASSIGN, loc}
	case "..":
		return token{RANGE_EXCLUSIVE, loc}
	case "..=":
		return token{RANGE_INCLUSIVE, loc}
	case "->":
		return token{SLIM_ARR, loc}
	case "=>":
		return token{FAT_ARR, loc}
	}
	switch {
	case number.MatchString(text):
		return literal{NUMBER, text, loc}
	case word.MatchString(text):
		return literal{IDENTIFIER, text, loc}
	}
	return token{ILLEGAL, loc}
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
