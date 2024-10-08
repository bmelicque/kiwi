package parser

import "testing"

func TestEmptyBrackets(t *testing.T) {
	tokenizer := testTokenizer{tokens: []Token{
		token{kind: LeftBracket},
		token{kind: RightBracket},
	}}
	parser := MakeParser(&tokenizer)
	parser.parseBracketedExpression()

	if len(parser.errors) != 0 {
		t.Fatalf("Expected no errors, got %v: %#v", len(parser.errors), parser.errors)
	}
}

func TestSimpleBracket(t *testing.T) {
	tokenizer := testTokenizer{tokens: []Token{
		token{kind: LeftBracket},
		literal{kind: Name, value: "Type"},
		token{kind: RightBracket},
	}}
	parser := MakeParser(&tokenizer)
	parser.parseBracketedExpression()

	if len(parser.errors) != 0 {
		t.Fatalf("Expected no errors, got %v: %#v", len(parser.errors), parser.errors)
	}
}

func TestBracketedTuple(t *testing.T) {
	tokenizer := testTokenizer{tokens: []Token{
		token{kind: LeftBracket},
		literal{kind: Name, value: "Type"},
		token{kind: Comma},
		literal{kind: Name, value: "Type"},
		token{kind: RightBracket},
	}}
	parser := MakeParser(&tokenizer)
	parser.parseBracketedExpression()

	if len(parser.errors) != 0 {
		t.Fatalf("Expected no errors, got %v: %#v", len(parser.errors), parser.errors)
	}
}
