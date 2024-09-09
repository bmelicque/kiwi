package checker

import (
	"reflect"
	"testing"

	"github.com/bmelicque/test-parser/parser"
	"github.com/bmelicque/test-parser/tokenizer"
)

func TestFunctionCall(t *testing.T) {
	checker := MakeChecker()
	checker.scope.Add("function", tokenizer.Loc{}, Function{[]Generic{}, Tuple{}, Primitive{NUMBER}})
	checker.checkCallExpression(parser.CallExpression{
		Callee: parser.TokenExpression{Token: testToken{tokenizer.IDENTIFIER, "function", tokenizer.Loc{}}},
		Args:   parser.ParenthesizedExpression{Expr: nil},
	})

	if len(checker.errors) != 0 {
		t.Fatalf("Expected no errors, got %#v", checker.errors)
	}
}

func TestGenericFunctionCall(t *testing.T) {
	checker := MakeChecker()
	checker.scope.Add(
		"function",
		tokenizer.Loc{},
		Function{
			[]Generic{{Name: "Type"}},
			Tuple{[]ExpressionType{Generic{Name: "Type"}}},
			Generic{Name: "Type"},
		},
	)
	expr := checker.checkCallExpression(parser.CallExpression{
		Callee: parser.TokenExpression{Token: testToken{tokenizer.IDENTIFIER, "function", tokenizer.Loc{}}},
		Args:   parser.ParenthesizedExpression{Expr: parser.TokenExpression{Token: testToken{tokenizer.NUMBER, "42", tokenizer.Loc{}}}},
	})

	if len(checker.errors) != 0 {
		t.Fatalf("Expected no errors, got %#v", checker.errors)
	}

	call, ok := expr.(CallExpression)
	if !ok {
		t.Fatalf("Expected CallExpression, got %v", reflect.TypeOf(expr))
	}

	if call.Typing.Kind() != NUMBER {
		t.Fatalf("Expected call to return NUMBER, got %#v", call.Typing)
	}
}
