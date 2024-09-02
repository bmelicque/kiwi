package checker

import (
	"testing"

	"github.com/bmelicque/test-parser/parser"
	"github.com/bmelicque/test-parser/tokenizer"
)

func TestFunctionCall(t *testing.T) {
	checker := MakeChecker()
	checker.scope.Add("function", tokenizer.Loc{}, Function{[]string{}, Tuple{}, Primitive{NUMBER}})
	checker.checkCallExpression(parser.CallExpression{
		Callee: parser.TokenExpression{Token: testToken{tokenizer.IDENTIFIER, "function", tokenizer.Loc{}}},
		Args:   &parser.ParenthesizedExpression{Expr: nil},
	})

	if len(checker.errors) != 0 {
		t.Fatalf("Expected no errors, got %#v", checker.errors)
	}
}
