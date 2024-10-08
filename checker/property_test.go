package checker

import (
	"testing"

	"github.com/bmelicque/test-parser/parser"
)

func TestSumTypeConstructor1(t *testing.T) {
	checker := MakeChecker()
	checker.scope.Add(
		"Sum",
		parser.Loc{},
		Type{TypeAlias{
			Name: "Sum",
			Ref: Sum{map[string]ExpressionType{
				"A": Type{Primitive{NUMBER}},
				"B": nil,
			}},
		}},
	)
	expr := checker.checkPropertyAccess(parser.PropertyAccessExpression{
		Expr:     parser.TokenExpression{Token: testToken{kind: parser.IDENTIFIER, value: "Sum"}},
		Property: parser.TokenExpression{Token: testToken{kind: parser.IDENTIFIER, value: "B"}},
	}).(PropertyAccessExpression)

	if len(checker.errors) != 0 {
		t.Fatalf("Expected no errors, got %#v", checker.errors)
	}

	alias, ok := expr.typing.(TypeAlias)
	if !ok {
		t.Fatalf("Expected alias, got %#v", expr.typing)
	}

	if _, ok = alias.Ref.(Sum); !ok {
		t.Fatalf("Expected sum, got %#v", alias)
	}
}

func TestSumTypeConstructor2(t *testing.T) {
	checker := MakeChecker()
	checker.scope.Add(
		"Sum",
		parser.Loc{},
		Type{TypeAlias{
			Name: "Sum",
			Ref: Sum{map[string]ExpressionType{
				"A": Type{Primitive{NUMBER}},
				"B": nil,
			}},
		}},
	)
	expr := checker.checkPropertyAccess(parser.PropertyAccessExpression{
		Expr:     parser.TokenExpression{Token: testToken{kind: parser.IDENTIFIER, value: "Sum"}},
		Property: parser.TokenExpression{Token: testToken{kind: parser.IDENTIFIER, value: "A"}},
	}).(PropertyAccessExpression)

	if len(checker.errors) != 0 {
		t.Fatalf("Expected no errors, got %#v", checker.errors)
	}

	typing, ok := expr.typing.(Type)
	if !ok {
		t.Fatalf("Expected type, got %#v", expr.typing)
	}

	if typing.Value.Kind() != CONSTRUCTOR {
		t.Fatalf("Expected constructor, got %#v", typing.Value)
	}
}

func TestTupleIndexAccess(t *testing.T) {
	checker := MakeChecker()
	checker.scope.Add(
		"tuple",
		parser.Loc{},
		Tuple{[]ExpressionType{Primitive{NUMBER}, Primitive{STRING}}},
	)
	expr := checker.checkPropertyAccess(parser.PropertyAccessExpression{
		Expr:     parser.TokenExpression{Token: testToken{kind: parser.IDENTIFIER, value: "tuple"}},
		Property: parser.TokenExpression{Token: testToken{kind: parser.NUMBER, value: "1"}},
	}).(PropertyAccessExpression)

	if len(checker.errors) != 0 {
		t.Fatalf("Expected no errors, got %#v", checker.errors)
	}

	if expr.typing.Kind() != STRING {
		t.Fatalf("Expected STRING, got %#v", expr.typing)
	}
}

func TestTraitExpression(t *testing.T) {
	checker := MakeChecker()
	expr := checker.checkPropertyAccess(parser.PropertyAccessExpression{
		Expr: parser.ParenthesizedExpression{Expr: parser.TokenExpression{Token: testToken{kind: parser.IDENTIFIER, value: "Self"}}},
		Property: parser.ParenthesizedExpression{Expr: parser.TypedExpression{
			Expr: parser.TokenExpression{Token: testToken{kind: parser.IDENTIFIER, value: "method"}},
			Typing: parser.FunctionExpression{
				Params:   &parser.ParenthesizedExpression{},
				Operator: testToken{kind: parser.SLIM_ARR},
				Expr:     parser.TokenExpression{Token: testToken{kind: parser.IDENTIFIER, value: "Self"}},
			},
		}},
	})

	if len(checker.errors) > 0 {
		t.Fatalf("Got %v checking errors: %#v", len(checker.errors), checker.errors)
	}

	trait, ok := expr.(TraitExpression)
	if !ok {
		t.Fatalf("Expected TraitExpression, got %#v", expr)
	}

	if _, ok := trait.Receiver.Expr.(Identifier); !ok {
		t.Fatalf("Expected Identifier, got %#v", trait.Receiver.Expr)
	}
}
