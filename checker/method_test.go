package checker

import (
	"testing"

	"github.com/bmelicque/test-parser/parser"
)

func TestMethodDeclaration(t *testing.T) {
	checker := MakeChecker()
	checker.scope.Add("Type", parser.Loc{}, Type{TypeAlias{
		Name: "Type",
		Ref:  Object{map[string]ExpressionType{"n": Type{Primitive{NUMBER}}}},
	}})
	// (t Type).method :: () ->
	checker.checkMethodDeclaration(parser.Assignment{
		Declared: parser.PropertyAccessExpression{
			Expr: parser.ParenthesizedExpression{Expr: parser.TypedExpression{
				Expr:   parser.TokenExpression{Token: testToken{parser.IDENTIFIER, "t", parser.Loc{}}},
				Typing: parser.TokenExpression{Token: testToken{parser.IDENTIFIER, "Type", parser.Loc{}}},
			}},
			Property: parser.TokenExpression{Token: testToken{parser.IDENTIFIER, "method", parser.Loc{}}},
		},
		Operator: testToken{parser.DEFINE, "::", parser.Loc{}},
		Initializer: parser.FunctionExpression{
			Params:   &parser.ParenthesizedExpression{},
			Operator: testToken{kind: parser.FAT_ARR},
			Body:     &parser.Block{},
		},
	})

	if len(checker.errors) != 1 {
		// only unused variable for receiver
		t.Fatalf("Expected 1 error, got %#v", checker.errors)
	}
}
