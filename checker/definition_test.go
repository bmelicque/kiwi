package checker

import (
	"testing"

	"github.com/bmelicque/test-parser/parser"
)

func TestObjectDeclaration(t *testing.T) {
	checker := MakeChecker()
	assignment := checker.checkDefinition(parser.Assignment{
		Declared: parser.TokenExpression{Token: testToken{kind: parser.IDENTIFIER, value: "Type"}},
		Operator: testToken{kind: parser.ASSIGN},
		Initializer: parser.ParenthesizedExpression{
			Expr: parser.TypedExpression{
				Expr:   parser.TokenExpression{Token: testToken{kind: parser.IDENTIFIER, value: "n"}},
				Typing: parser.TokenExpression{Token: testToken{kind: parser.NUM_KW}},
			},
		},
	})

	if len(checker.errors) != 0 {
		t.Fatalf("Expected no errors, got %#v", checker.errors)
	}

	declaration, ok := assignment.(VariableDeclaration)
	if !ok {
		t.Fatalf("Expected VariableDeclaration, got %#v", assignment)
	}

	if _, ok := declaration.Pattern.(Identifier); !ok {
		t.Fatalf("Expected identifier 'n', got %#v", declaration.Pattern)
	}

	variable, _ := checker.scope.Find("Type")
	if variable == nil {
		t.Fatalf("Expected type to be added to scope")
		return
	}
	typing, ok := variable.typing.(Type)
	if !ok {
		t.Fatalf("Expected 'Type' type")
	}
	if _, ok := typing.Value.(TypeAlias); !ok {
		t.Fatalf("Expected 'TypeAlias' subtype, got %#v", typing.Value)
	}
}

func TestGenericObjectDefinition(t *testing.T) {
	checker := MakeChecker()
	checker.checkDefinition(parser.Assignment{
		Declared: parser.ComputedAccessExpression{
			Expr:     parser.TokenExpression{Token: testToken{kind: parser.IDENTIFIER, value: "Generic"}},
			Property: parser.BracketedExpression{Expr: parser.TokenExpression{Token: testToken{kind: parser.IDENTIFIER, value: "TypeParam"}}},
		},
		Operator: testToken{parser.ASSIGN, "::", parser.Loc{}},
		Initializer: parser.ParenthesizedExpression{
			Expr: parser.TypedExpression{
				Expr:   parser.TokenExpression{Token: testToken{kind: parser.IDENTIFIER, value: "value"}},
				Typing: parser.TokenExpression{Token: testToken{kind: parser.IDENTIFIER, value: "TypeParam"}},
			},
		},
	})

	if len(checker.errors) > 0 {
		t.Fatalf("Expected no errors, got %v: %#v", len(checker.errors), checker.errors)
	}
}
