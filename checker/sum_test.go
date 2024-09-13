package checker

import (
	"testing"

	"github.com/bmelicque/test-parser/parser"
	"github.com/bmelicque/test-parser/tokenizer"
)

func TestSumTypeDefinition(t *testing.T) {
	checker := MakeChecker()
	sum := checker.checkSumType(parser.SumType{
		Members: []parser.Node{
			parser.TypedExpression{
				Expr:   parser.TokenExpression{Token: testToken{tokenizer.IDENTIFIER, "Some", tokenizer.Loc{}}},
				Typing: parser.TokenExpression{Token: testToken{tokenizer.NUM_KW, "number", tokenizer.Loc{}}},
			},
			parser.TokenExpression{Token: testToken{tokenizer.IDENTIFIER, "None", tokenizer.Loc{}}},
		},
	})

	if len(checker.errors) != 0 {
		t.Fatalf("Expected no errors, got %#v", checker.errors)
	}

	if len(sum.Members) != 2 {
		t.Fatalf("Expected 2 members, got %v", len(sum.Members))
	}

	member := sum.Members[0]
	if _, ok := member.Typing.(Literal); !ok {
		t.Fatalf("Expected literal type 'number', got %#v", member.Typing)
	}
}
