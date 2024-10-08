package parser

import "testing"

func TestEmptyBlock(t *testing.T) {
	tokenizer := testTokenizer{tokens: []Token{
		token{kind: LeftBrace},
		token{kind: RightBrace},
	}}
	parser := MakeParser(&tokenizer)
	parser.parseBlock()

	if len(parser.errors) != 0 {
		t.Fatalf("Expected no errors, got %#v", parser.errors)
	}
}

func TestSingleLineBlock(t *testing.T) {
	tokenizer := testTokenizer{tokens: []Token{
		token{kind: LeftBrace},
		literal{kind: StringLiteral, value: "Hello, world!"},
		token{kind: RightBrace},
	}}
	parser := MakeParser(&tokenizer)
	parser.parseBlock()

	if len(parser.errors) != 0 {
		t.Fatalf("Expected no errors, got %#v", parser.errors)
	}
}

func TestMultilineBlock(t *testing.T) {
	tokenizer := testTokenizer{tokens: []Token{
		token{kind: LeftBrace},
		token{kind: EOL},

		literal{kind: StringLiteral, value: "Hello, world!"},
		token{kind: EOL},

		literal{kind: StringLiteral, value: "Hello, world!"},
		token{kind: EOL},

		token{kind: RightBrace},
		token{kind: EOL},
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
