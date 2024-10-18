package checker

import (
	"testing"

	"github.com/bmelicque/test-parser/parser"
)

func TestTupleDeclaration(t *testing.T) {
	checker := MakeChecker()
	assignment := checker.checkVariableDeclaration(parser.Assignment{
		Pattern: parser.TupleExpression{
			Elements: []parser.Node{
				parser.TokenExpression{Token: testToken{parser.Name, "n", parser.Loc{}}},
				parser.TokenExpression{Token: testToken{parser.Name, "s", parser.Loc{}}},
			},
		},
		Value: parser.TupleExpression{
			Elements: []parser.Node{
				parser.TokenExpression{Token: testToken{parser.NumberLiteral, "1", parser.Loc{}}},
				parser.TokenExpression{Token: testToken{parser.StringLiteral, "\"string\"", parser.Loc{}}},
			},
		},
		Operator: testToken{parser.Assign, ":=", parser.Loc{}},
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
