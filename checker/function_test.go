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
			Expr:   parser.TokenExpression{Token: testToken{kind: tokenizer.IDENTIFIER, value: "n"}},
			Typing: parser.TokenExpression{Token: testToken{kind: tokenizer.NUM_KW}},
		}},
		Operator: testToken{kind: tokenizer.SLIM_ARR},
		Expr: parser.BinaryExpression{
			Right:    parser.TokenExpression{Token: testToken{kind: tokenizer.NUMBER, value: "2"}},
			Left:     parser.TokenExpression{Token: testToken{kind: tokenizer.IDENTIFIER, value: "n"}},
			Operator: testToken{kind: tokenizer.MUL},
		},
	})

	if len(checker.errors) != 0 {
		t.Fatalf("Expected no errors, got %#v", checker.errors)
	}

	if e, ok := expr.(SlimArrowFunction); !ok {
		t.Fatalf("Expected SlimArrowFunction, got %#v", e)
	}
}

func TestGenericFunctionExpression(t *testing.T) {
	checker := MakeChecker()
	expr := checker.checkFunctionExpression(parser.FunctionExpression{
		TypeParams: &parser.BracketedExpression{Expr: parser.TokenExpression{Token: testToken{kind: tokenizer.IDENTIFIER, value: "Type"}}},
		Params: &parser.ParenthesizedExpression{Expr: parser.TypedExpression{
			Expr:   parser.TokenExpression{Token: testToken{kind: tokenizer.IDENTIFIER, value: "value"}},
			Typing: parser.TokenExpression{Token: testToken{kind: tokenizer.IDENTIFIER, value: "Type"}},
		}},
		Operator: testToken{kind: tokenizer.SLIM_ARR},
		Expr:     parser.TokenExpression{Token: testToken{kind: tokenizer.IDENTIFIER, value: "value"}},
	})

	if len(checker.errors) != 0 {
		t.Fatalf("Expected no errors, got %#v", checker.errors)
	}

	if e, ok := expr.(SlimArrowFunction); !ok {
		t.Fatalf("Expected SlimArrowFunction, got %#v", e)
	}
}

func TestFunctionType(t *testing.T) {
	checker := MakeChecker()
	expr := checker.checkFunctionExpression(parser.FunctionExpression{
		Params:   &parser.ParenthesizedExpression{Expr: parser.TokenExpression{Token: testToken{kind: tokenizer.NUM_KW}}},
		Operator: testToken{kind: tokenizer.SLIM_ARR},
		Expr:     parser.TokenExpression{Token: testToken{kind: tokenizer.NUM_KW}},
	})

	if len(checker.errors) != 0 {
		t.Fatalf("Expected no errors, got %#v", checker.errors)
	}

	if _, ok := expr.(FunctionTypeExpression); !ok {
		t.Fatalf("Expected FunctionTypeExpression, got %#v", expr)
	}
}
