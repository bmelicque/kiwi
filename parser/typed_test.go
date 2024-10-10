package parser

import "testing"

func TestTypedExpression(t *testing.T) {
	tokenizer := testTokenizer{tokens: []Token{
		literal{kind: Name, value: "n"},
		token{kind: NumberKeyword},
	}}
	parser := MakeParser(&tokenizer)
	node := parser.parseTypedExpression()

	_, ok := node.(*TypedExpression)
	if !ok {
		t.Fatalf("Expected TypedExpression, got %#v", node)
	}
}

func TestTypedExpressionWithColon(t *testing.T) {
	tokenizer := testTokenizer{tokens: []Token{
		literal{kind: Name, value: "n"},
		token{kind: Colon},
		token{kind: NumberKeyword},
	}}
	parser := MakeParser(&tokenizer)
	node := parser.parseTypedExpression()

	_, ok := node.(*TypedExpression)
	if !ok {
		t.Fatalf("Expected TypedExpression, got %#v", node)
	}
	if len(parser.errors) != 0 {
		t.Fatalf("Expected no errors, got %+v: %#v", len(parser.errors), parser.errors)
	}
}
