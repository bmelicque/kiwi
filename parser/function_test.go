package parser

import (
	"testing"

	"github.com/bmelicque/test-parser/tokenizer"
)

func TestSlimArrowFunctionWithoutArgs(t *testing.T) {
	parser := MakeParser(&testTokenizer{tokens: []tokenizer.Token{
		testToken{kind: tokenizer.LPAREN},
		testToken{kind: tokenizer.RPAREN},
		testToken{kind: tokenizer.SLIM_ARR},
		testToken{kind: tokenizer.NUMBER, value: "42"},
	}})
	node := parser.parseFunctionExpression()

	function, ok := node.(FunctionExpression)
	if !ok {
		t.Fatalf("Expected FunctionExpression, got %#v", node)
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
	if function.Body != nil {
		t.Fatalf("Expected no Body, got %#v", function.Body)
	}
}

func TestSlimArrowFunctionWithArgs(t *testing.T) {
	parser := MakeParser(&testTokenizer{tokens: []tokenizer.Token{
		testToken{kind: tokenizer.LPAREN},
		testToken{kind: tokenizer.IDENTIFIER, value: "n"},
		testToken{kind: tokenizer.NUM_KW},
		testToken{kind: tokenizer.RPAREN},
		testToken{kind: tokenizer.SLIM_ARR},
		testToken{kind: tokenizer.NUMBER, value: "2"},
		testToken{kind: tokenizer.MUL},
		testToken{kind: tokenizer.IDENTIFIER, value: "n"},
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
	if function.Body != nil {
		t.Fatalf("Expected no Body, got %#v", function.Body)
	}
}

func TestFunctionType(t *testing.T) {
	// (number) -> number
	parser := MakeParser(&testTokenizer{tokens: []tokenizer.Token{
		testToken{kind: tokenizer.LPAREN},
		testToken{kind: tokenizer.NUM_KW},
		testToken{kind: tokenizer.RPAREN},
		testToken{kind: tokenizer.SLIM_ARR},
		testToken{kind: tokenizer.NUM_KW},
	}})
	node := parser.parseFunctionExpression()

	function, ok := node.(FunctionExpression)
	if !ok {
		t.Fatalf("Expected FunctionExpression, got %#v", node)
	}

	if function.Operator.Kind() != tokenizer.SLIM_ARR {
		t.Fatalf("Expected '->', got %v", function.Operator.Text())
	}
	if _, ok := function.Expr.(TokenExpression); !ok {
		t.Fatalf("Expected TokenExpression, got %#v", function.Expr)
	}
	if function.Body != nil {
		t.Fatalf("Expected no Body, got %#v", function.Body)
	}
}

func TestFatArrowFunctionWithoutArgs(t *testing.T) {
	parser := MakeParser(&testTokenizer{tokens: []tokenizer.Token{
		testToken{kind: tokenizer.LPAREN},
		testToken{kind: tokenizer.RPAREN},
		testToken{kind: tokenizer.FAT_ARR},
		testToken{kind: tokenizer.NUM_KW},
		testToken{kind: tokenizer.LBRACE},
		testToken{kind: tokenizer.RETURN_KW},
		testToken{kind: tokenizer.NUMBER, value: "42"},
		testToken{kind: tokenizer.RBRACE},
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
	if function.Operator.Kind() != tokenizer.FAT_ARR {
		t.Fatalf("Expected '=>', got %v", function.Operator.Text())
	}
	if _, ok := function.Expr.(TokenExpression); !ok {
		t.Fatalf("Expected TokenExpression, got %#v", function.Expr)
	}
	if function.Body == nil {
		t.Fatalf("Expected Body, got nothing")
	}
}

func TestFatArrowFunctionWithArgs(t *testing.T) {
	parser := MakeParser(&testTokenizer{tokens: []tokenizer.Token{
		testToken{kind: tokenizer.LPAREN},
		testToken{kind: tokenizer.IDENTIFIER, value: "n"},
		testToken{kind: tokenizer.NUM_KW},
		testToken{kind: tokenizer.RPAREN},
		testToken{kind: tokenizer.FAT_ARR},
		testToken{kind: tokenizer.NUM_KW},
		testToken{kind: tokenizer.LBRACE},
		testToken{kind: tokenizer.RETURN_KW},
		testToken{kind: tokenizer.IDENTIFIER, value: "n"},
		testToken{kind: tokenizer.RBRACE},
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
	if function.Operator.Kind() != tokenizer.FAT_ARR {
		t.Fatalf("Expected '=>', got %v", function.Operator.Text())
	}
	if _, ok := function.Expr.(TokenExpression); !ok {
		t.Fatalf("Expected TokenExpression, got %#v", function.Expr)
	}
	if function.Body == nil {
		t.Fatalf("Expected Body, got nothing")
	}
}

func TestFunctionWithTypeArgs(t *testing.T) {
	parser := MakeParser(&testTokenizer{tokens: []tokenizer.Token{
		testToken{kind: tokenizer.LBRACKET},
		testToken{kind: tokenizer.IDENTIFIER, value: "Type"},
		testToken{kind: tokenizer.RBRACKET},
		testToken{kind: tokenizer.LPAREN},
		testToken{kind: tokenizer.RPAREN},
		testToken{kind: tokenizer.FAT_ARR},
		testToken{kind: tokenizer.LBRACE},
		testToken{kind: tokenizer.RBRACE},
	}})
	node := parser.parseFunctionExpression()

	_, ok := node.(FunctionExpression)
	if !ok {
		t.Fatalf("Expected FunctionExpression, got %#v", node)
		return
	}
}
