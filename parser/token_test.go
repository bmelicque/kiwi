package parser

import (
	"testing"

	"github.com/bmelicque/test-parser/tokenizer"
)

func TestToken(t *testing.T) {
	boolean := testToken{
		kind:  tokenizer.BOOLEAN,
		value: "true",
		loc:   tokenizer.Loc{},
	}
	tokenizer := testTokenizer{tokens: []tokenizer.Token{boolean}}
	parser := MakeParser(&tokenizer)
	node := parser.parseTokenExpression()
	if _, ok := node.(TokenExpression); !ok {
		t.Fatalf("Expected TokenExpression, got %#v", node)
	}
}
