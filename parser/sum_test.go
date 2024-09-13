package parser

import (
	"testing"

	"github.com/bmelicque/test-parser/tokenizer"
)

func TestSumType(t *testing.T) {
	tokenizer := testTokenizer{tokens: []tokenizer.Token{
		testToken{tokenizer.BOR, "|", tokenizer.Loc{}},
		testToken{tokenizer.IDENTIFIER, "Some", tokenizer.Loc{}},
		testToken{tokenizer.IDENTIFIER, "Type", tokenizer.Loc{}},
		testToken{tokenizer.BOR, "|", tokenizer.Loc{}},
		testToken{tokenizer.IDENTIFIER, "None", tokenizer.Loc{}},
	}}
	parser := MakeParser(&tokenizer)
	node := parser.parseSumType()

	if len(parser.errors) > 0 {
		t.Fatalf("Expected no errors, got %v: %#v", len(parser.errors), parser.errors)
	}

	sum, ok := node.(SumType)
	if !ok {
		t.Fatalf("Expected SumType, got %#v", node)
		return
	}
	if len(sum.Members) != 2 {
		t.Fatalf("Expected 2 elements, got %v: %#v", len(sum.Members), sum.Members)
	}
}
