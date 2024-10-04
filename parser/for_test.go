package parser

import (
	"testing"

	"github.com/bmelicque/test-parser/tokenizer"
)

func TestForExpression(t *testing.T) {
	tok := testTokenizer{tokens: []tokenizer.Token{
		testToken{kind: tokenizer.FOR_KW},
		testToken{kind: tokenizer.BOOLEAN, value: "true"},
		testToken{kind: tokenizer.LBRACE},
		testToken{kind: tokenizer.NUMBER, value: "42"},
		testToken{kind: tokenizer.RBRACE},
	}}
	parser := MakeParser(&tok)
	parser.parseForExpression()

	if len(parser.errors) != 0 {
		t.Fatalf("Expected no errors, got %#v", parser.errors)
	}
}
