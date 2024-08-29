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
}
