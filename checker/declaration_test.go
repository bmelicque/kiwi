package checker

import (
	"testing"

	"github.com/bmelicque/test-parser/parser"
)

func TestSimpleDeclaration(t *testing.T) {
	checker := MakeChecker()
	assignment := checker.checkVariableDeclaration(parser.Assignment{
		Declared:    parser.TokenExpression{Token: testToken{parser.IDENTIFIER, "n", parser.Loc{}}},
		Initializer: parser.TokenExpression{Token: testToken{parser.NUMBER, "42", parser.Loc{}}},
		Operator:    testToken{parser.ASSIGN, ":=", parser.Loc{}},
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
				parser.TokenExpression{Token: testToken{parser.IDENTIFIER, "n", parser.Loc{}}},
				parser.TokenExpression{Token: testToken{parser.IDENTIFIER, "s", parser.Loc{}}},
			},
		},
		Initializer: parser.TupleExpression{
			Elements: []parser.Node{
				parser.TokenExpression{Token: testToken{parser.NUMBER, "1", parser.Loc{}}},
				parser.TokenExpression{Token: testToken{parser.STRING, "\"string\"", parser.Loc{}}},
			},
		},
		Operator: testToken{parser.ASSIGN, ":=", parser.Loc{}},
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
