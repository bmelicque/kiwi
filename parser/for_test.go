package parser

import "testing"

func TestForExpression(t *testing.T) {
	// for true { 42 }
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

func TestForExpressionType(t *testing.T) {
	// for true { break 42 }
	tok := testTokenizer{tokens: []Token{
		token{kind: ForKeyword},
		literal{kind: BooleanLiteral, value: "true"},
		token{kind: LeftBrace},
		token{kind: BreakKeyword},
		literal{kind: NumberLiteral, value: "42"},
		token{kind: RightBrace},
	}}
	parser := MakeParser(&tok)
	expr := parser.parseForExpression()

	if len(parser.errors) != 0 {
		t.Fatalf("Expected no errors, got %#v", parser.errors)
	}
	if expr.Type().Kind() != NUMBER {
		t.Fatalf("Expected number type, got %#v", expr.Type())
	}
}
