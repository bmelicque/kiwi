package parser

import "testing"

func TestReturn(t *testing.T) {
	tok := testTokenizer{tokens: []Token{
		token{kind: ReturnKeyword},
		literal{kind: BooleanLiteral, value: "true"},
	}}
	parser := MakeParser(&tok)
	exit := parser.parseExit()

	if len(parser.errors) != 0 {
		t.Fatalf("Expected no errors, got %#v", parser.errors)
	}
	if exit.Operator.Kind() != ReturnKeyword {
		t.Fatal("Expected 'return' keyword")
	}
}

func TestBreak(t *testing.T) {
	tok := testTokenizer{tokens: []Token{
		token{kind: BreakKeyword},
		literal{kind: BooleanLiteral, value: "true"},
	}}
	parser := MakeParser(&tok)
	exit := parser.parseExit()

	if len(parser.errors) != 0 {
		t.Fatalf("Expected no errors, got %#v", parser.errors)
	}
	if exit.Operator.Kind() != BreakKeyword {
		t.Fatal("Expected 'return' keyword")
	}
}

func TestContinue(t *testing.T) {
	tok := testTokenizer{tokens: []Token{
		token{kind: ContinueKeyword},
	}}
	parser := MakeParser(&tok)
	exit := parser.parseExit()

	if len(parser.errors) != 0 {
		t.Fatalf("Expected no errors, got %#v", parser.errors)
	}
	if exit.Operator.Kind() != ContinueKeyword {
		t.Fatal("Expected 'return' keyword")
	}
}
