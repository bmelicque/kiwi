package parser

import (
	"testing"

	"github.com/bmelicque/test-parser/tokenizer"
)

func TestIf(t *testing.T) {
	tokenizer := testTokenizer{tokens: []tokenizer.Token{
		testToken{tokenizer.IF_KW, "if", tokenizer.Loc{}},
		testToken{tokenizer.IDENTIFIER, "n", tokenizer.Loc{}},
		testToken{tokenizer.EQ, "==", tokenizer.Loc{}},
		testToken{tokenizer.NUMBER, "2", tokenizer.Loc{}},
		testToken{tokenizer.LBRACE, "{", tokenizer.Loc{}},
		testToken{tokenizer.RETURN_KW, "return", tokenizer.Loc{}},
		testToken{tokenizer.IDENTIFIER, "1", tokenizer.Loc{}},
		testToken{tokenizer.RBRACE, "}", tokenizer.Loc{}},
	}}
	parser := MakeParser(&tokenizer)
	node := parser.parseIf()
	statement, ok := node.(IfElse)
	if !ok {
		t.Fatalf("Expected 'if' statement, got %#v", node)
		return
	}
	if statement.Body == nil {
		t.Fatalf("Expected 'body' statement, got %#v", node)
	}
}
