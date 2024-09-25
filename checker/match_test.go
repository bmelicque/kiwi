package checker

import (
	"testing"

	"github.com/bmelicque/test-parser/parser"
	"github.com/bmelicque/test-parser/tokenizer"
)

func TestMatchStatement(t *testing.T) {
	checker := MakeChecker()
	typing := TypeAlias{
		Name: "Option",
		Ref: Sum{map[string]ExpressionType{
			"Some": Type{Primitive{NUMBER}},
			"None": nil,
		}},
	}
	checker.scope.Add("Option", tokenizer.Loc{}, Type{typing})
	checker.scope.Add("option", tokenizer.Loc{}, typing)
	// match option {
	// case s Some:
	//		s
	// case None:
	// }
	expr := checker.checkMatchStatement(parser.MatchStatement{
		Keyword: testToken{kind: tokenizer.MATCH_KW},
		Value:   parser.TokenExpression{Token: testToken{kind: tokenizer.IDENTIFIER, value: "option"}},
		Cases: []parser.MatchCase{
			{
				Pattern: parser.TypedExpression{
					Expr:   parser.TokenExpression{Token: testToken{kind: tokenizer.IDENTIFIER, value: "s"}},
					Typing: parser.TokenExpression{Token: testToken{kind: tokenizer.IDENTIFIER, value: "Some"}},
				},
				Statements: []parser.Node{
					parser.ExpressionStatement{Expr: parser.TokenExpression{Token: testToken{kind: tokenizer.IDENTIFIER, value: "s"}}},
				},
			},
			{
				Pattern:    parser.TokenExpression{Token: testToken{kind: tokenizer.IDENTIFIER, value: "None"}},
				Statements: []parser.Node{},
			},
		},
	})

	if len(checker.errors) != 0 {
		t.Fatalf("Expected no errors, got %#v", checker.errors)
	}
	_ = expr
}
