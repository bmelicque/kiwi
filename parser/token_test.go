package parser

import "testing"

func TestToken(t *testing.T) {
	tokenizer := testTokenizer{tokens: []Token{
		literal{kind: BooleanLiteral, value: "true"},
	}}
	parser := MakeParser(&tokenizer)
	node := parser.parseTokenExpression()
	if _, ok := node.(TokenExpression); !ok {
		t.Fatalf("Expected TokenExpression, got %#v", node)
	}
}
