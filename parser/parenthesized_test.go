package parser

import (
	"testing"

	"github.com/bmelicque/test-parser/tokenizer"
)

func TestParenthesized(t *testing.T) {
	tokenizer := testTokenizer{tokens: []tokenizer.Token{
		testToken{tokenizer.LPAREN, "(", tokenizer.Loc{}},
		testToken{tokenizer.NUMBER, "42", tokenizer.Loc{}},
		testToken{tokenizer.RPAREN, ")", tokenizer.Loc{}},
	}}
	parser := MakeParser(&tokenizer)
	paren := parser.parseParenthesizedExpression()
	if _, ok := paren.Expr.(TokenExpression); !ok {
		t.Fatalf("Expected TokenExpression between parentheses, got %v", paren.Expr)
	}
}
