package checker

import (
	"reflect"
	"testing"

	"github.com/bmelicque/test-parser/parser"
)

func TestFunctionCall(t *testing.T) {
	checker := MakeChecker()
	checker.scope.Add("function", parser.Loc{}, Function{[]Generic{}, Tuple{}, Primitive{NUMBER}})
	expr := checker.checkCallExpression(parser.CallExpression{
		Callee: parser.TokenExpression{Token: testToken{parser.IDENTIFIER, "function", parser.Loc{}}},
		Args:   parser.ParenthesizedExpression{Expr: nil},
	})

	if len(checker.errors) != 0 {
		t.Fatalf("Expected no errors, got %#v", checker.errors)
	}

	if expr.Type().Kind() != NUMBER {
		t.Fatalf("Expected type to be number, got %#v", expr.Type())
	}
}

func TestGenericFunctionCall(t *testing.T) {
	checker := MakeChecker()
	checker.scope.Add(
		"function",
		parser.Loc{},
		Function{
			[]Generic{{Name: "Type"}},
			Tuple{[]ExpressionType{Generic{Name: "Type"}}},
			Generic{Name: "Type"},
		},
	)
	expr := checker.checkCallExpression(parser.CallExpression{
		Callee: parser.TokenExpression{Token: testToken{kind: parser.IDENTIFIER, value: "function"}},
		Args: parser.ParenthesizedExpression{
			Expr: parser.TokenExpression{Token: testToken{kind: parser.NUMBER, value: "42"}},
		},
	})

	if len(checker.errors) != 0 {
		t.Fatalf("Expected no errors, got %#v", checker.errors)
	}

	call, ok := expr.(CallExpression)
	if !ok {
		t.Fatalf("Expected CallExpression, got %v", reflect.TypeOf(expr))
	}

	if call.typing.Kind() != NUMBER {
		t.Fatalf("Expected call to return NUMBER, got %#v", call.typing)
	}
}
