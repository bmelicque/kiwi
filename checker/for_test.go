package checker

import (
	"testing"

	"github.com/bmelicque/test-parser/parser"
	"github.com/bmelicque/test-parser/tokenizer"
)

func TestForLoop(t *testing.T) {
	checker := MakeChecker()
	loop := checker.checkLoop(parser.ForExpression{
		Statement: parser.TokenExpression{
			Token: testToken{kind: tokenizer.BOOLEAN, value: "true"},
		},
		Body: &parser.Block{Statements: []parser.Node{
			parser.IfElse{
				Condition: parser.TokenExpression{
					Token: testToken{kind: tokenizer.BOOLEAN, value: "true"},
				},
				Body: &parser.Block{Statements: []parser.Node{
					parser.Exit{
						Operator: testToken{kind: tokenizer.BREAK_KW},
						Value: parser.TokenExpression{
							Token: testToken{kind: tokenizer.NUMBER, value: "42"},
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
