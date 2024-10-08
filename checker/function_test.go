package checker

import (
	"testing"

	"github.com/bmelicque/test-parser/parser"
)

func TestCheckFunctionReturns(t *testing.T) {
	checker := MakeChecker()
	body := Block{Statements: []Node{
		If{
			Condition: Literal{TokenExpression: parser.TokenExpression{
				Token: testToken{kind: parser.BOOLEAN, value: "true"},
			}},
			Block: Block{Statements: []Node{
				Exit{
					Operator: testToken{kind: parser.RETURN_KW},
					Value: Literal{TokenExpression: parser.TokenExpression{
						Token: testToken{kind: parser.NUMBER, value: "42"},
					}},
				},
			}},
		},
		Literal{TokenExpression: parser.TokenExpression{
			Token: testToken{kind: parser.NUMBER, value: "42"},
		}},
	}}
	checkFunctionReturns(checker, body)
	if len(checker.errors) > 0 {
		t.Fatalf("Expected no errors, got %#v", checker.errors)
	}
}

func TestGenericFunctionExpression(t *testing.T) {
	checker := MakeChecker()
	expr := checker.checkFunctionExpression(parser.FunctionExpression{
		TypeParams: &parser.BracketedExpression{Expr: parser.TokenExpression{Token: testToken{kind: parser.IDENTIFIER, value: "Type"}}},
		Params: &parser.ParenthesizedExpression{Expr: parser.TypedExpression{
			Expr:   parser.TokenExpression{Token: testToken{kind: parser.IDENTIFIER, value: "value"}},
			Typing: parser.TokenExpression{Token: testToken{kind: parser.IDENTIFIER, value: "Type"}},
		}},
		Operator: testToken{kind: parser.FAT_ARR},
		Body: &parser.Block{Statements: []parser.Node{
			parser.TokenExpression{Token: testToken{kind: parser.IDENTIFIER, value: "value"}},
		}},
	})

	if len(checker.errors) != 0 {
		t.Fatalf("Expected no errors, got %#v", checker.errors)
	}

	if e, ok := expr.(FunctionExpression); !ok {
		t.Fatalf("Expected FunctionExpression, got %#v", e)
	}
}

func TestFunctionType(t *testing.T) {
	checker := MakeChecker()
	expr := checker.checkFunctionExpression(parser.FunctionExpression{
		Params:   &parser.ParenthesizedExpression{Expr: parser.TokenExpression{Token: testToken{kind: parser.NUM_KW}}},
		Operator: testToken{kind: parser.SLIM_ARR},
		Expr:     parser.TokenExpression{Token: testToken{kind: parser.NUM_KW}},
	})

	if len(checker.errors) != 0 {
		t.Fatalf("Expected no errors, got %#v", checker.errors)
	}

	if _, ok := expr.(FunctionTypeExpression); !ok {
		t.Fatalf("Expected FunctionTypeExpression, got %#v", expr)
	}
}
