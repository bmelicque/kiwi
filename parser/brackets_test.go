package parser

import "testing"

func TestEmptyBrackets(t *testing.T) {
	tokenizer := testTokenizer{tokens: []Token{
		token{kind: LBRACKET},
		token{kind: RBRACKET},
	}}
	parser := MakeParser(&tokenizer)
	parser.parseBracketedExpression()

	if len(parser.errors) != 0 {
		t.Fatalf("Expected no errors, got %v: %#v", len(parser.errors), parser.errors)
	}
}

func TestSimpleBracket(t *testing.T) {
	tokenizer := testTokenizer{tokens: []Token{
		token{kind: LBRACKET},
		literal{kind: IDENTIFIER, value: "Type"},
		token{kind: RBRACKET},
	}}
	parser := MakeParser(&tokenizer)
	parser.parseBracketedExpression()

	if len(parser.errors) != 0 {
		t.Fatalf("Expected no errors, got %v: %#v", len(parser.errors), parser.errors)
	}
}

func TestBracketedTuple(t *testing.T) {
	tokenizer := testTokenizer{tokens: []Token{
		token{kind: LBRACKET},
		literal{kind: IDENTIFIER, value: "Type"},
		token{kind: COMMA},
		literal{kind: IDENTIFIER, value: "Type"},
		token{kind: RBRACKET},
	}}
	parser := MakeParser(&tokenizer)
	parser.parseBracketedExpression()

	if len(parser.errors) != 0 {
		t.Fatalf("Expected no errors, got %v: %#v", len(parser.errors), parser.errors)
	}
}
