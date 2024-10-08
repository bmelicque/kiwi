package parser

import "testing"

func TestForExpression(t *testing.T) {
	tok := testTokenizer{tokens: []Token{
		token{kind: ForKeyword},
		literal{kind: BooleanLiteral, value: "true"},
		token{kind: LeftBrace},
		literal{kind: NumberLiteral, value: "42"},
		token{kind: RightBrace},
	}}
	parser := MakeParser(&tok)
	parser.parseForExpression()

	if len(parser.errors) != 0 {
		t.Fatalf("Expected no errors, got %#v", parser.errors)
	}
}
