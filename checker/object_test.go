package checker

import (
	"testing"

	"github.com/bmelicque/test-parser/parser"
	"github.com/bmelicque/test-parser/tokenizer"
)

func TestObjectExpression(t *testing.T) {
	checker := MakeChecker()
	checker.scope.Add("Type", tokenizer.Loc{}, Type{TypeAlias{Name: "Type", Ref: Object{map[string]ExpressionType{"n": Type{Primitive{NUMBER}}}}}})
	checker.checkObjectExpression(parser.ObjectExpression{
		Typing: parser.TokenExpression{Token: testToken{kind: tokenizer.IDENTIFIER, value: "Type"}},
		Members: []parser.Node{
			parser.TypedExpression{
				Expr:   parser.TokenExpression{Token: testToken{tokenizer.IDENTIFIER, "n", tokenizer.Loc{}}},
				Typing: parser.TokenExpression{Token: testToken{tokenizer.NUMBER, "42", tokenizer.Loc{}}},
				Colon:  true,
			},
		},
	})

	if len(checker.errors) != 0 {
		t.Fatalf("Expected no errors, got %#v", checker.errors)
	}
}

func TestObjectExpressionColons(t *testing.T) {
	checker := MakeChecker()
	checker.scope.Add("Type", tokenizer.Loc{}, Type{TypeAlias{Name: "Type", Ref: Object{map[string]ExpressionType{"n": Type{Primitive{NUMBER}}}}}})
	checker.checkObjectExpression(parser.ObjectExpression{
		Typing: parser.TokenExpression{Token: testToken{kind: tokenizer.IDENTIFIER, value: "Type"}},
		Members: []parser.Node{
			parser.TypedExpression{
				Expr:   parser.TokenExpression{Token: testToken{tokenizer.IDENTIFIER, "n", tokenizer.Loc{}}},
				Typing: parser.TokenExpression{Token: testToken{tokenizer.NUMBER, "42", tokenizer.Loc{}}},
				Colon:  false,
			},
		},
	})

	if len(checker.errors) != 1 {
		t.Fatalf("Expected 1 error, got %v, %#v", len(checker.errors), checker.errors)
	}
}
