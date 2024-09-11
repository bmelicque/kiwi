package checker

import (
	"testing"

	"github.com/bmelicque/test-parser/parser"
	"github.com/bmelicque/test-parser/tokenizer"
)

func TestObjectExpression(t *testing.T) {
	checker := MakeChecker()
	checker.scope.Add("Type", tokenizer.Loc{}, Type{TypeAlias{Name: "Type", Ref: Object{map[string]ExpressionType{"n": Type{Primitive{NUMBER}}}}}})
	checker.checkInstanciationExpression(parser.InstanciationExpression{
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
	checker.checkInstanciationExpression(parser.InstanciationExpression{
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

func TestGenericObjectExpression(t *testing.T) {
	checker := MakeChecker()
	checker.scope.Add(
		"Generic",
		tokenizer.Loc{},
		Type{TypeAlias{
			Name:   "Generic",
			Params: []Generic{{Name: "Type"}},
			Ref:    Object{map[string]ExpressionType{"value": Type{Generic{Name: "Type"}}}},
		}},
	)
	checker.checkInstanciationExpression(parser.InstanciationExpression{
		Typing: parser.TokenExpression{Token: testToken{tokenizer.IDENTIFIER, "Generic", tokenizer.Loc{Start: tokenizer.Position{Col: 0}}}},
		Members: []parser.Node{
			parser.TypedExpression{
				Expr:   parser.TokenExpression{Token: testToken{tokenizer.IDENTIFIER, "value", tokenizer.Loc{Start: tokenizer.Position{Col: 1}}}},
				Typing: parser.TokenExpression{Token: testToken{tokenizer.NUMBER, "42", tokenizer.Loc{Start: tokenizer.Position{Col: 2}}}},
				Colon:  true,
			},
		},
	})

	if len(checker.errors) != 0 {
		t.Fatalf("Expected no errors, got %#v", checker.errors)
	}
}

func TestListExpression(t *testing.T) {
	checker := MakeChecker()
	checker.checkInstanciationExpression(parser.InstanciationExpression{
		Typing: parser.ListTypeExpression{
			Bracketed: parser.BracketedExpression{},
			Type:      parser.TokenExpression{Token: testToken{kind: tokenizer.NUM_KW, value: "number"}},
		},
		Members: []parser.Node{
			parser.TokenExpression{Token: testToken{kind: tokenizer.NUMBER, value: "1"}},
			parser.TokenExpression{Token: testToken{kind: tokenizer.NUMBER, value: "2"}},
		},
	})

	if len(checker.errors) != 0 {
		t.Fatalf("Expected no errors, got %#v", checker.errors)
	}
}
