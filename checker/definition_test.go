package checker

import (
	"testing"

	"github.com/bmelicque/test-parser/parser"
	"github.com/bmelicque/test-parser/tokenizer"
)

func TestObjectDeclaration(t *testing.T) {
	checker := MakeChecker()
	assignment := checker.checkDefinition(parser.Assignment{
		Declared: parser.TokenExpression{Token: testToken{tokenizer.IDENTIFIER, "Type", tokenizer.Loc{}}},
		Operator: testToken{tokenizer.ASSIGN, "::", tokenizer.Loc{}},
		Initializer: parser.ObjectDefinition{
			Members: []parser.Node{
				parser.TypedExpression{
					Expr:   parser.TokenExpression{Token: testToken{tokenizer.IDENTIFIER, "n", tokenizer.Loc{}}},
					Typing: parser.TokenExpression{Token: testToken{tokenizer.NUM_KW, "number", tokenizer.Loc{}}},
				},
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
	if _, ok := declaration.Initializer.(ObjectDefinition); !ok {
		t.Fatalf("Expected ObjectDefinition, got %#v", declaration.Initializer)
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
			Expr:     parser.TokenExpression{Token: testToken{kind: tokenizer.IDENTIFIER, value: "Generic"}},
			Property: parser.BracketedExpression{Expr: parser.TokenExpression{Token: testToken{kind: tokenizer.IDENTIFIER, value: "TypeParam"}}},
		},
		Operator: testToken{tokenizer.ASSIGN, "::", tokenizer.Loc{}},
		Initializer: parser.ObjectDefinition{
			Members: []parser.Node{
				parser.TypedExpression{
					Expr:   parser.TokenExpression{Token: testToken{kind: tokenizer.IDENTIFIER, value: "value"}},
					Typing: parser.TokenExpression{Token: testToken{kind: tokenizer.IDENTIFIER, value: "TypeParam"}},
				},
			},
		},
	})

	if len(checker.errors) > 0 {
		t.Fatalf("Expected no errors, got %v: %#v", len(checker.errors), checker.errors)
	}
}
