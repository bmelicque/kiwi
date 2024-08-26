package checker

import (
	"testing"

	"github.com/bmelicque/test-parser/parser"
	"github.com/bmelicque/test-parser/tokenizer"
)

func TestFunctionGenericType(t *testing.T) {
	checker := MakeChecker()
	expr := checker.checkFunctionExpression(parser.FunctionExpression{
		TypeParams: &parser.AngleExpression{Expr: parser.TokenExpression{Token: testToken{tokenizer.IDENTIFIER, "Type", tokenizer.Loc{}}}},
		Operator:   testToken{tokenizer.SLIM_ARR, "->", tokenizer.Loc{}},
		Expr: parser.ObjectDefinition{Members: []parser.Node{
			parser.TypedExpression{
				Expr:   parser.TokenExpression{Token: testToken{tokenizer.IDENTIFIER, "value", tokenizer.Loc{}}},
				Typing: parser.TokenExpression{Token: testToken{tokenizer.IDENTIFIER, "Type", tokenizer.Loc{}}},
			},
		}},
	})

	if len(checker.errors) != 0 {
		t.Fatalf("Expected no errors, got %#v", checker.errors)
	}

	if e, ok := expr.(GenericTypeDef); !ok {
		t.Fatalf("Expected GenericTypeDef, got %#v", e)
	}
}
