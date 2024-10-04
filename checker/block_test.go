package checker

import (
	"testing"

	"github.com/bmelicque/test-parser/parser"
	"github.com/bmelicque/test-parser/tokenizer"
)

func TestBlockType(t *testing.T) {
	checker := MakeChecker()
	block := checker.checkBlock(parser.Block{
		Statements: []parser.Node{
			parser.TokenExpression{Token: testToken{kind: tokenizer.STRING, value: "\"Hello, world!\""}},
			parser.TokenExpression{Token: testToken{kind: tokenizer.NUMBER, value: "42"}},
		},
	})

	if len(checker.errors) != 0 {
		t.Fatalf("Expected no errors, got %#v", checker.errors)
	}
	if block.Type().Kind() != NUMBER {
		t.Fatalf("Expected number type, got %#v", block.Type())
	}
}

func TestUnreachableCode(t *testing.T) {
	checker := MakeChecker()
	checker.checkBlock(parser.Block{
		Statements: []parser.Node{
			parser.Exit{Operator: testToken{kind: tokenizer.RETURN_KW}},
			parser.TokenExpression{Token: testToken{
				kind:  tokenizer.STRING,
				value: "\"Hello, world!\"",
				loc:   tokenizer.Loc{Start: tokenizer.Position{Col: 1}},
			}},
		},
	})

	if len(checker.errors) != 2 {
		// also get one error for returning while not being in a function
		t.Fatalf("Expected 2 errors, got %#v", checker.errors)
	}
}
