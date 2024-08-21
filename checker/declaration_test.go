package checker

import (
	"testing"

	"github.com/bmelicque/test-parser/parser"
	"github.com/bmelicque/test-parser/tokenizer"
)

func TestSimpleDeclaration(t *testing.T) {
	checker := MakeChecker()
	assignment := checker.checkVariableDeclaration(parser.Assignment{
		Declared:    parser.TokenExpression{Token: testToken{tokenizer.IDENTIFIER, "n", tokenizer.Loc{}}},
		Initializer: parser.TokenExpression{Token: testToken{tokenizer.NUMBER, "42", tokenizer.Loc{}}},
		Operator:    testToken{tokenizer.ASSIGN, ":=", tokenizer.Loc{}},
	})

	if len(checker.errors) != 0 {
		t.Fatalf("Expected no errors, got %#v", checker.errors)
	}
	if _, ok := assignment.Pattern.(Identifier); !ok {
		t.Fatalf("Expected identifier 'n', got %#v", assignment.Pattern)
	}
	if _, ok := assignment.Initializer.(Literal); !ok {
		t.Fatalf("Expected literal 42, got %#v", assignment.Initializer)
	}
}

func TestTupleDeclaration(t *testing.T) {
	checker := MakeChecker()
	assignment := checker.checkVariableDeclaration(parser.Assignment{
		Declared: parser.TupleExpression{
			Elements: []parser.Node{
				parser.TokenExpression{Token: testToken{tokenizer.IDENTIFIER, "n", tokenizer.Loc{}}},
				parser.TokenExpression{Token: testToken{tokenizer.IDENTIFIER, "s", tokenizer.Loc{}}},
			},
		},
		Initializer: parser.TupleExpression{
			Elements: []parser.Node{
				parser.TokenExpression{Token: testToken{tokenizer.NUMBER, "1", tokenizer.Loc{}}},
				parser.TokenExpression{Token: testToken{tokenizer.STRING, "\"string\"", tokenizer.Loc{}}},
			},
		},
		Operator: testToken{tokenizer.ASSIGN, ":=", tokenizer.Loc{}},
	})

	if len(checker.errors) != 0 {
		t.Fatalf("Expected no errors, got %#v", checker.errors)
	}
	if _, ok := assignment.Pattern.(TupleExpression); !ok {
		t.Fatalf("Expected identifier tuple, got %#v", assignment.Pattern)
	}
	if _, ok := assignment.Initializer.(TupleExpression); !ok {
		t.Fatalf("Expected literal tuple, got %#v", assignment.Initializer)
	}
}

func TestObjectDeclaration(t *testing.T) {
	checker := MakeChecker()
	assignment := checker.checkVariableDeclaration(parser.Assignment{
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
	if _, ok := assignment.Pattern.(Identifier); !ok {
		t.Fatalf("Expected identifier 'n', got %#v", assignment.Pattern)
	}
	if _, ok := assignment.Initializer.(ObjectDefinition); !ok {
		t.Fatalf("Expected ObjectDefinition, got %#v", assignment.Initializer)
	}
}
