package checker

import (
	"testing"

	"github.com/bmelicque/test-parser/parser"
)

func TestMatchExpression(t *testing.T) {
	checker := MakeChecker()
	typing := makeOptionType(Primitive{NUMBER})
	checker.scope.Add("Option", parser.Loc{}, Type{typing})
	checker.scope.Add("option", parser.Loc{}, typing)
	// match option {
	// case s Some:
	//		s
	// case None:
	//		0
	// }
	expr := checker.checkMatchExpression(parser.MatchExpression{
		Keyword: testToken{kind: parser.MatchKeyword},
		Value:   parser.TokenExpression{Token: testToken{kind: parser.Name, value: "option"}},
		Cases: []parser.MatchCase{
			{
				Pattern: parser.TypedExpression{
					Expr:   parser.TokenExpression{Token: testToken{kind: parser.Name, value: "s"}},
					Typing: parser.TokenExpression{Token: testToken{kind: parser.Name, value: "Some"}},
				},
				Statements: []parser.Node{
					parser.ExpressionStatement{Expr: parser.TokenExpression{Token: testToken{kind: parser.Name, value: "s"}}},
				},
			},
			{
				Pattern: parser.TokenExpression{Token: testToken{kind: parser.Name, value: "None"}},
				Statements: []parser.Node{
					parser.ExpressionStatement{Expr: parser.TokenExpression{Token: testToken{kind: parser.NumberLiteral, value: "0"}}},
				},
			},
		},
	})

	if len(checker.errors) != 0 {
		t.Fatalf("Expected no errors, got %#v", checker.errors)
	}
	if expr.Type().Kind() != NUMBER {
		t.Fatalf("Expected number type, got %#v", expr.Type())
	}
}
