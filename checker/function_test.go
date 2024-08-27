package checker

import (
	"testing"

	"github.com/bmelicque/test-parser/parser"
	"github.com/bmelicque/test-parser/tokenizer"
)

func TestSlimArrowFunction(t *testing.T) {
	checker := MakeChecker()
	expr := checker.checkFunctionExpression(parser.FunctionExpression{
		Params: &parser.ParenthesizedExpression{Expr: parser.TypedExpression{
			Expr:   parser.TokenExpression{Token: testToken{tokenizer.IDENTIFIER, "n", tokenizer.Loc{}}},
			Typing: parser.TokenExpression{Token: testToken{tokenizer.NUM_KW, "number", tokenizer.Loc{}}},
		}},
		Operator: testToken{tokenizer.SLIM_ARR, "->", tokenizer.Loc{}},
		Expr: parser.BinaryExpression{
			Right:    parser.TokenExpression{Token: testToken{tokenizer.NUMBER, "2", tokenizer.Loc{}}},
			Left:     parser.TokenExpression{Token: testToken{tokenizer.IDENTIFIER, "n", tokenizer.Loc{}}},
			Operator: testToken{tokenizer.MUL, "*", tokenizer.Loc{}},
		},
	})

	if len(checker.errors) != 0 {
		t.Fatalf("Expected no errors, got %#v", checker.errors)
	}

	if e, ok := expr.(SlimArrowFunction); !ok {
		t.Fatalf("Expected SlimArrowFunction, got %#v", e)
	}
}

func TestGenericTypeDeclaration(t *testing.T) {
	checker := MakeChecker()
	expr := checker.checkFunctionExpression(parser.FunctionExpression{
		TypeParams: &parser.AngleExpression{Expr: parser.TokenExpression{Token: testToken{tokenizer.IDENTIFIER, "Type", tokenizer.Loc{}}}},
		Operator:   testToken{tokenizer.SLIM_ARR, "->", tokenizer.Loc{}},
		Expr: parser.ObjectDefinition{Members: []parser.Node{
			parser.TypedExpression{
				Expr:   parser.TokenExpression{Token: testToken{tokenizer.IDENTIFIER, "value", tokenizer.Loc{}}},
				Typing: parser.TokenExpression{Token: testToken{tokenizer.IDENTIFIER, "Type", tokenizer.Loc{}}},
			},
		}},
	})

	if len(checker.errors) != 0 {
		t.Fatalf("Expected no errors, got %#v", checker.errors)
	}

	if e, ok := expr.(GenericTypeDef); !ok {
		t.Fatalf("Expected GenericTypeDef, got %#v", e)
	}
}

func TestGenericFunctionExpression(t *testing.T) {
	checker := MakeChecker()
	expr := checker.checkFunctionExpression(parser.FunctionExpression{
		TypeParams: &parser.AngleExpression{Expr: parser.TokenExpression{Token: testToken{tokenizer.IDENTIFIER, "Type", tokenizer.Loc{}}}},
		Params: &parser.ParenthesizedExpression{Expr: parser.TypedExpression{
			Expr:   parser.TokenExpression{Token: testToken{tokenizer.IDENTIFIER, "value", tokenizer.Loc{}}},
			Typing: parser.TokenExpression{Token: testToken{tokenizer.IDENTIFIER, "Type", tokenizer.Loc{}}},
		}},
		Operator: testToken{tokenizer.SLIM_ARR, "->", tokenizer.Loc{}},
		Expr:     parser.TokenExpression{Token: testToken{tokenizer.IDENTIFIER, "value", tokenizer.Loc{}}},
	})

	if len(checker.errors) != 0 {
		t.Fatalf("Expected no errors, got %#v", checker.errors)
	}

	if e, ok := expr.(SlimArrowFunction); !ok {
		t.Fatalf("Expected SlimArrowFunction, got %#v", e)
	}
}
