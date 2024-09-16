package checker

import (
	"testing"

	"github.com/bmelicque/test-parser/parser"
	"github.com/bmelicque/test-parser/tokenizer"
)

func TestSumTypeConstructor1(t *testing.T) {
	checker := MakeChecker()
	checker.scope.Add(
		"Sum",
		tokenizer.Loc{},
		Type{TypeAlias{
			Name: "Sum",
			Ref: Sum{map[string]ExpressionType{
				"A": Type{Primitive{NUMBER}},
				"B": nil,
			}},
		}},
	)
	expr := checker.checkPropertyAccess(parser.PropertyAccessExpression{
		Expr:     parser.TokenExpression{Token: testToken{kind: tokenizer.IDENTIFIER, value: "Sum"}},
		Property: parser.TokenExpression{Token: testToken{kind: tokenizer.IDENTIFIER, value: "B"}},
	})

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
		tokenizer.Loc{},
		Type{TypeAlias{
			Name: "Sum",
			Ref: Sum{map[string]ExpressionType{
				"A": Type{Primitive{NUMBER}},
				"B": nil,
			}},
		}},
	)
	expr := checker.checkPropertyAccess(parser.PropertyAccessExpression{
		Expr:     parser.TokenExpression{Token: testToken{kind: tokenizer.IDENTIFIER, value: "Sum"}},
		Property: parser.TokenExpression{Token: testToken{kind: tokenizer.IDENTIFIER, value: "A"}},
	})

	if len(checker.errors) != 0 {
		t.Fatalf("Expected no errors, got %#v", checker.errors)
	}

	typing, ok := expr.typing.(Type)
	if !ok {
		t.Fatalf("Expected type, got %#v", expr.typing)
	}

	if typing.Value.Kind() != NUMBER {
		t.Fatalf("Expected number, got %#v", typing.Value)
	}
}
