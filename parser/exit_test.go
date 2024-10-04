package parser

import (
	"testing"

	"github.com/bmelicque/test-parser/tokenizer"
)

func TestReturn(t *testing.T) {
	tok := testTokenizer{tokens: []tokenizer.Token{
		testToken{kind: tokenizer.RETURN_KW},
		testToken{kind: tokenizer.BOOLEAN, value: "true"},
	}}
	parser := MakeParser(&tok)
	exit := parser.parseExit()

	if len(parser.errors) != 0 {
		t.Fatalf("Expected no errors, got %#v", parser.errors)
	}
	if exit.Operator.Kind() != tokenizer.RETURN_KW {
		t.Fatal("Expected 'return' keyword")
	}
}

func TestBreak(t *testing.T) {
	tok := testTokenizer{tokens: []tokenizer.Token{
		testToken{kind: tokenizer.BREAK_KW},
		testToken{kind: tokenizer.BOOLEAN, value: "true"},
	}}
	parser := MakeParser(&tok)
	exit := parser.parseExit()

	if len(parser.errors) != 0 {
		t.Fatalf("Expected no errors, got %#v", parser.errors)
	}
	if exit.Operator.Kind() != tokenizer.BREAK_KW {
		t.Fatal("Expected 'return' keyword")
	}
}

func TestContinue(t *testing.T) {
	tok := testTokenizer{tokens: []tokenizer.Token{
		testToken{kind: tokenizer.CONTINUE_KW},
	}}
	parser := MakeParser(&tok)
	exit := parser.parseExit()

	if len(parser.errors) != 0 {
		t.Fatalf("Expected no errors, got %#v", parser.errors)
	}
	if exit.Operator.Kind() != tokenizer.CONTINUE_KW {
		t.Fatal("Expected 'return' keyword")
	}
}
