package parser

import (
	"testing"

	"github.com/bmelicque/test-parser/tokenizer"
)

func TestEmptyBlock(t *testing.T) {
	tokenizer := testTokenizer{tokens: []tokenizer.Token{
		testToken{kind: tokenizer.LBRACE},
		testToken{kind: tokenizer.RBRACE},
	}}
	parser := MakeParser(&tokenizer)
	parser.parseBlock()

	if len(parser.errors) != 0 {
		t.Fatalf("Expected no errors, got %#v", parser.errors)
	}
}

func TestSingleLineBlock(t *testing.T) {
	tokenizer := testTokenizer{tokens: []tokenizer.Token{
		testToken{kind: tokenizer.LBRACE},
		testToken{kind: tokenizer.STRING, value: "Hello, world!"},
		testToken{kind: tokenizer.RBRACE},
	}}
	parser := MakeParser(&tokenizer)
	parser.parseBlock()

	if len(parser.errors) != 0 {
		t.Fatalf("Expected no errors, got %#v", parser.errors)
	}
}

func TestMultilineBlock(t *testing.T) {
	tokenizer := testTokenizer{tokens: []tokenizer.Token{
		testToken{kind: tokenizer.LBRACE},
		testToken{kind: tokenizer.EOL},

		testToken{kind: tokenizer.STRING, value: "Hello, world!"},
		testToken{kind: tokenizer.EOL},

		testToken{kind: tokenizer.STRING, value: "Hello, world!"},
		testToken{kind: tokenizer.EOL},

		testToken{kind: tokenizer.RBRACE},
		testToken{kind: tokenizer.EOL},
	}}
	parser := MakeParser(&tokenizer)
	block := parser.parseBlock()

	if len(parser.errors) != 0 {
		t.Fatalf("Expected no errors, got %#v", parser.errors)
	}
	if len(block.Statements) != 2 {
		t.Fatalf("Expected 2 statements, got %#v", block.Statements)
	}
}
