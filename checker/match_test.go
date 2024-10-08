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
		Keyword: testToken{kind: parser.MATCH_KW},
		Value:   parser.TokenExpression{Token: testToken{kind: parser.IDENTIFIER, value: "option"}},
		Cases: []parser.MatchCase{
			{
				Pattern: parser.TypedExpression{
					Expr:   parser.TokenExpression{Token: testToken{kind: parser.IDENTIFIER, value: "s"}},
					Typing: parser.TokenExpression{Token: testToken{kind: parser.IDENTIFIER, value: "Some"}},
				},
				Statements: []parser.Node{
					parser.ExpressionStatement{Expr: parser.TokenExpression{Token: testToken{kind: parser.IDENTIFIER, value: "s"}}},
				},
			},
			{
				Pattern: parser.TokenExpression{Token: testToken{kind: parser.IDENTIFIER, value: "None"}},
				Statements: []parser.Node{
					parser.ExpressionStatement{Expr: parser.TokenExpression{Token: testToken{kind: parser.NUMBER, value: "0"}}},
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
