package parser

import (
	"testing"

	"github.com/bmelicque/test-parser/tokenizer"
)

func TestListTypeExpression(t *testing.T) {
	tokenizer := testTokenizer{tokens: []tokenizer.Token{
		testToken{tokenizer.LBRACKET, "[", tokenizer.Loc{}},
		testToken{tokenizer.RBRACKET, "]", tokenizer.Loc{}},
		testToken{tokenizer.NUM_KW, "number", tokenizer.Loc{}},
	}}
	parser := MakeParser(&tokenizer)
	node := ParseExpression(parser)

	if len(parser.errors) != 0 {
		t.Fatalf("Expected no errors, got %+v: %#v", len(parser.errors), parser.errors)
	}

	list, ok := node.(ListTypeExpression)
	if !ok {
		t.Fatalf("Expected ListExpression, got %#v", node)
	}
	if list.Type == nil {
		t.Fatalf("Expected a Type")
	}
}
