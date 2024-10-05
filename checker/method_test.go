package checker

import (
	"testing"

	"github.com/bmelicque/test-parser/parser"
	"github.com/bmelicque/test-parser/tokenizer"
)

func TestMethodDeclaration(t *testing.T) {
	checker := MakeChecker()
	checker.scope.Add("Type", tokenizer.Loc{}, Type{TypeAlias{
		Name: "Type",
		Ref:  Object{map[string]ExpressionType{"n": Type{Primitive{NUMBER}}}},
	}})
	// (t Type).method :: () ->
	checker.checkMethodDeclaration(parser.Assignment{
		Declared: parser.PropertyAccessExpression{
			Expr: parser.ParenthesizedExpression{Expr: parser.TypedExpression{
				Expr:   parser.TokenExpression{Token: testToken{tokenizer.IDENTIFIER, "t", tokenizer.Loc{}}},
				Typing: parser.TokenExpression{Token: testToken{tokenizer.IDENTIFIER, "Type", tokenizer.Loc{}}},
			}},
			Property: parser.TokenExpression{Token: testToken{tokenizer.IDENTIFIER, "method", tokenizer.Loc{}}},
		},
		Operator: testToken{tokenizer.DEFINE, "::", tokenizer.Loc{}},
		Initializer: parser.FunctionExpression{
			Params:   &parser.ParenthesizedExpression{},
			Operator: testToken{kind: tokenizer.FAT_ARR},
			Body:     &parser.Block{},
		},
	})

	if len(checker.errors) != 1 {
		// only unused variable for receiver
		t.Fatalf("Expected 1 error, got %#v", checker.errors)
	}
}
