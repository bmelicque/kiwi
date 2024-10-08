package parser

import "testing"

func TestLiteral(t *testing.T) {
	parser := MakeParser(&testTokenizer{tokens: []Token{
		literal{kind: BooleanLiteral, value: "true"},
	}})
	expr := parser.parseToken(false)
	if _, ok := expr.(*Literal); !ok {
		t.Fatalf("Expected TokenExpression, got %#v", expr)
	}
	if expr.Type().Kind() != BOOLEAN {
		t.Fatalf("Expected boolean, got %#v", expr.Type())
	}
}
