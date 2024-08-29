package parser

import (
	"testing"

	"github.com/bmelicque/test-parser/tokenizer"
)

func TestEmptyBrackets(t *testing.T) {
	tokenizer := testTokenizer{tokens: []tokenizer.Token{
		testToken{tokenizer.LBRACKET, "[", tokenizer.Loc{}},
		testToken{tokenizer.RBRACKET, "]", tokenizer.Loc{}},
	}}
	parser := MakeParser(&tokenizer)
	parser.parseBracketedExpression()

	if len(parser.errors) != 0 {
		t.Fatalf("Expected no errors, got %v: %#v", len(parser.errors), parser.errors)
	}
}

func TestSimpleBracket(t *testing.T) {
	tokenizer := testTokenizer{tokens: []tokenizer.Token{
		testToken{tokenizer.LBRACKET, "[", tokenizer.Loc{}},
		testToken{tokenizer.IDENTIFIER, "Type", tokenizer.Loc{}},
		testToken{tokenizer.RBRACKET, "]", tokenizer.Loc{}},
	}}
	parser := MakeParser(&tokenizer)
	parser.parseBracketedExpression()

	if len(parser.errors) != 0 {
		t.Fatalf("Expected no errors, got %v: %#v", len(parser.errors), parser.errors)
	}
}

func TestBracketedTuple(t *testing.T) {
	tokenizer := testTokenizer{tokens: []tokenizer.Token{
		testToken{tokenizer.LBRACKET, "[", tokenizer.Loc{}},
		testToken{tokenizer.IDENTIFIER, "Type", tokenizer.Loc{}},
		testToken{tokenizer.COMMA, ",", tokenizer.Loc{}},
		testToken{tokenizer.IDENTIFIER, "Type", tokenizer.Loc{}},
		testToken{tokenizer.RBRACKET, "]", tokenizer.Loc{}},
	}}
	parser := MakeParser(&tokenizer)
	parser.parseBracketedExpression()

	if len(parser.errors) != 0 {
		t.Fatalf("Expected no errors, got %v: %#v", len(parser.errors), parser.errors)
	}
}
