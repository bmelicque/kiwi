package parser

import "testing"

func TestReturn(t *testing.T) {
	tok := testTokenizer{tokens: []Token{
		token{kind: RETURN_KW},
		literal{kind: BOOLEAN, value: "true"},
	}}
	parser := MakeParser(&tok)
	exit := parser.parseExit()

	if len(parser.errors) != 0 {
		t.Fatalf("Expected no errors, got %#v", parser.errors)
	}
	if exit.Operator.Kind() != RETURN_KW {
		t.Fatal("Expected 'return' keyword")
	}
}

func TestBreak(t *testing.T) {
	tok := testTokenizer{tokens: []Token{
		token{kind: BREAK_KW},
		literal{kind: BOOLEAN, value: "true"},
	}}
	parser := MakeParser(&tok)
	exit := parser.parseExit()

	if len(parser.errors) != 0 {
		t.Fatalf("Expected no errors, got %#v", parser.errors)
	}
	if exit.Operator.Kind() != BREAK_KW {
		t.Fatal("Expected 'return' keyword")
	}
}

func TestContinue(t *testing.T) {
	tok := testTokenizer{tokens: []Token{
		token{kind: CONTINUE_KW},
	}}
	parser := MakeParser(&tok)
	exit := parser.parseExit()

	if len(parser.errors) != 0 {
		t.Fatalf("Expected no errors, got %#v", parser.errors)
	}
	if exit.Operator.Kind() != CONTINUE_KW {
		t.Fatal("Expected 'return' keyword")
	}
}
