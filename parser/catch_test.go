package parser

import "testing"

func TestParseCatchExpression(t *testing.T) {
	parser := MakeParser(&testTokenizer{tokens: []Token{
		literal{kind: Name, value: "result"},
		token{kind: CatchKeyword},
		literal{kind: Name, value: "err"},
		token{kind: LeftBrace},
		literal{kind: BooleanLiteral, value: "true"},
		token{kind: RightBrace},
	}})
	expr := parser.parseCatchExpression()

	if len(parser.errors) > 0 {
		t.Fatalf("Expected no errors, got %#v", parser.errors)
	}
	if _, ok := expr.(*CatchExpression); !ok {
		t.Fatalf("Expected 'catch' expression, got %#v", expr)
	}
}

func TestParseCatchExpressionNoIdentifier(t *testing.T) {
	parser := MakeParser(&testTokenizer{tokens: []Token{
		literal{kind: Name, value: "result"},
		token{kind: CatchKeyword},
		token{kind: LeftBrace},
		token{kind: RightBrace},
	}})
	parser.parseCatchExpression()

	if len(parser.errors) != 1 {
		t.Fatalf("Expected 1 error, got %v: %#v", len(parser.errors), parser.errors)
	}
}

func TestParseCatchExpressionBadTokens(t *testing.T) {
	parser := MakeParser(&testTokenizer{tokens: []Token{
		literal{kind: Name, value: "result"},
		token{kind: CatchKeyword},
		literal{kind: Name, value: "err"},
		literal{kind: Name, value: "err"},
		token{kind: LeftBrace},
		token{kind: RightBrace},
	}})
	parser.parseCatchExpression()

	if len(parser.errors) != 1 {
		t.Fatalf("Expected 1 error, got %v: %#v", len(parser.errors), parser.errors)
	}
}
