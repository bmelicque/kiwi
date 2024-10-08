package parser

import "testing"

func TestForExpression(t *testing.T) {
	tok := testTokenizer{tokens: []Token{
		token{kind: FOR_KW},
		literal{kind: BOOLEAN, value: "true"},
		token{kind: LBRACE},
		literal{kind: NUMBER, value: "42"},
		token{kind: RBRACE},
	}}
	parser := MakeParser(&tok)
	parser.parseForExpression()

	if len(parser.errors) != 0 {
		t.Fatalf("Expected no errors, got %#v", parser.errors)
	}
}
