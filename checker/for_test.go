package checker

import (
	"testing"

	"github.com/bmelicque/test-parser/parser"
)

func TestForLoop(t *testing.T) {
	checker := MakeChecker()
	loop := checker.checkLoop(parser.ForExpression{
		Statement: parser.TokenExpression{
			Token: testToken{kind: parser.BooleanLiteral, value: "true"},
		},
		Body: &parser.Block{Statements: []parser.Node{
			parser.IfExpression{
				Condition: parser.TokenExpression{
					Token: testToken{kind: parser.BooleanLiteral, value: "true"},
				},
				Body: &parser.Block{Statements: []parser.Node{
					parser.Exit{
						Operator: testToken{kind: parser.BreakKeyword},
						Value: parser.TokenExpression{
							Token: testToken{kind: parser.NumberLiteral, value: "42"},
						},
					},
				}},
			},
		}},
	})

	if len(checker.errors) != 0 {
		t.Fatalf("Expected no errors, got %#v", checker.errors)
	}
	if loop.Type().Kind() != NUMBER {
		t.Fatalf("Expected number type, got %#v", loop.Type())
	}
}
