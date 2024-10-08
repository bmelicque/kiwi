package checker

import (
	"testing"

	"github.com/bmelicque/test-parser/parser"
)

func TestForLoop(t *testing.T) {
	checker := MakeChecker()
	loop := checker.checkLoop(parser.ForExpression{
		Statement: parser.TokenExpression{
			Token: testToken{kind: parser.BOOLEAN, value: "true"},
		},
		Body: &parser.Block{Statements: []parser.Node{
			parser.IfElse{
				Condition: parser.TokenExpression{
					Token: testToken{kind: parser.BOOLEAN, value: "true"},
				},
				Body: &parser.Block{Statements: []parser.Node{
					parser.Exit{
						Operator: testToken{kind: parser.BREAK_KW},
						Value: parser.TokenExpression{
							Token: testToken{kind: parser.NUMBER, value: "42"},
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
