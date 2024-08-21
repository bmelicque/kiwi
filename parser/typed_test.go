package parser

import (
	"testing"

	"github.com/bmelicque/test-parser/tokenizer"
)

func TestTypedExpression(t *testing.T) {
	tokenizer := testTokenizer{tokens: []tokenizer.Token{
		testToken{tokenizer.IDENTIFIER, "n", tokenizer.Loc{}},
		testToken{tokenizer.NUM_KW, "number", tokenizer.Loc{}},
	}}
	parser := MakeParser(&tokenizer)
	node := parser.parseTypedExpression()

	_, ok := node.(TypedExpression)
	if !ok {
		t.Fatalf("Expected TypedExpression, got %#v", node)
	}
}

func TestTypedExpressionWithColon(t *testing.T) {
	tokenizer := testTokenizer{tokens: []tokenizer.Token{
		testToken{tokenizer.IDENTIFIER, "n", tokenizer.Loc{}},
		testToken{tokenizer.COLON, ":", tokenizer.Loc{}},
		testToken{tokenizer.NUM_KW, "number", tokenizer.Loc{}},
	}}
	parser := MakeParser(&tokenizer)
	node := parser.parseTypedExpression()

	_, ok := node.(TypedExpression)
	if !ok {
		t.Fatalf("Expected TypedExpression, got %#v", node)
	}
	if len(parser.errors) != 0 {
		t.Fatalf("Expected no errors, got %+v: %#v", len(parser.errors), parser.errors)
	}
}
