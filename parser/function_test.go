package parser

import (
	"testing"

	"github.com/bmelicque/test-parser/tokenizer"
)

func TestSlimArrowFunctionWithoutArgs(t *testing.T) {
	parser := MakeParser(&testTokenizer{tokens: []tokenizer.Token{
		testToken{tokenizer.LPAREN, "(", tokenizer.Loc{}},
		testToken{tokenizer.RPAREN, ")", tokenizer.Loc{}},
		testToken{tokenizer.SLIM_ARR, "->", tokenizer.Loc{}},
		testToken{tokenizer.NUMBER, "42", tokenizer.Loc{}},
	}})
	node := parser.parseFunctionExpression()

	function, ok := node.(FunctionExpression)
	if !ok {
		t.Fatalf("Expected FunctionExpression, got %#v", node)
		return
	}

	if function.Params.Expr != nil {
		t.Fatalf("Expected no params, got %#v", function.Params.Expr)
	}
	if function.Operator.Kind() != tokenizer.SLIM_ARR {
		t.Fatalf("Expected '->', got %v", function.Operator.Text())
	}
	if _, ok := function.Expr.(TokenExpression); !ok {
		t.Fatalf("Expected TokenExpression, got %#v", function.Expr)
	}
}

func TestSlimArrowFunctionWithArgs(t *testing.T) {
	parser := MakeParser(&testTokenizer{tokens: []tokenizer.Token{
		testToken{tokenizer.LPAREN, "(", tokenizer.Loc{}},
		testToken{tokenizer.IDENTIFIER, "n", tokenizer.Loc{}},
		testToken{tokenizer.COLON, ":", tokenizer.Loc{}},
		testToken{tokenizer.NUM_KW, "number", tokenizer.Loc{}},
		testToken{tokenizer.RPAREN, ")", tokenizer.Loc{}},
		testToken{tokenizer.SLIM_ARR, "->", tokenizer.Loc{}},
		testToken{tokenizer.NUMBER, "2", tokenizer.Loc{}},
		testToken{tokenizer.MUL, "*", tokenizer.Loc{}},
		testToken{tokenizer.IDENTIFIER, "n", tokenizer.Loc{}},
	}})
	node := parser.parseFunctionExpression()

	function, ok := node.(FunctionExpression)
	if !ok {
		t.Fatalf("Expected FunctionExpression, got %#v", node)
		return
	}

	if _, ok := function.Params.Expr.(TypedExpression); !ok {
		t.Fatalf("Expected TypedExpression, got %#v", function.Params.Expr)
	}
	if function.Operator.Kind() != tokenizer.SLIM_ARR {
		t.Fatalf("Expected '->', got %v", function.Operator.Text())
	}
	if _, ok := function.Expr.(BinaryExpression); !ok {
		t.Fatalf("Expected BinaryExpression, got %#v", function.Expr)
	}
}
