package parser

import "testing"

func TestSlimArrowFunctionWithoutArgs(t *testing.T) {
	parser := MakeParser(&testTokenizer{tokens: []Token{
		token{kind: LPAREN},
		token{kind: RPAREN},
		token{kind: SLIM_ARR},
		literal{kind: NUMBER, value: "42"},
	}})
	node := parser.parseFunctionExpression()

	function, ok := node.(FunctionExpression)
	if !ok {
		t.Fatalf("Expected FunctionExpression, got %#v", node)
	}

	if function.Params.Expr != nil {
		t.Fatalf("Expected no params, got %#v", function.Params.Expr)
	}
	if function.Operator.Kind() != SLIM_ARR {
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
	parser := MakeParser(&testTokenizer{tokens: []Token{
		token{kind: LPAREN},
		literal{kind: IDENTIFIER, value: "n"},
		token{kind: NUM_KW},
		token{kind: RPAREN},
		token{kind: SLIM_ARR},
		literal{kind: NUMBER, value: "2"},
		token{kind: MUL},
		literal{kind: IDENTIFIER, value: "n"},
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
	if function.Operator.Kind() != SLIM_ARR {
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
	parser := MakeParser(&testTokenizer{tokens: []Token{
		token{kind: LPAREN},
		token{kind: NUM_KW},
		token{kind: RPAREN},
		token{kind: SLIM_ARR},
		token{kind: NUM_KW},
	}})
	node := parser.parseFunctionExpression()

	function, ok := node.(FunctionExpression)
	if !ok {
		t.Fatalf("Expected FunctionExpression, got %#v", node)
	}

	if function.Operator.Kind() != SLIM_ARR {
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
	parser := MakeParser(&testTokenizer{tokens: []Token{
		token{kind: LPAREN},
		token{kind: RPAREN},
		token{kind: FAT_ARR},
		token{kind: NUM_KW},
		token{kind: LBRACE},
		token{kind: RETURN_KW},
		literal{kind: NUMBER, value: "42"},
		token{kind: RBRACE},
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
	if function.Operator.Kind() != FAT_ARR {
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
	parser := MakeParser(&testTokenizer{tokens: []Token{
		token{kind: LPAREN},
		literal{kind: IDENTIFIER, value: "n"},
		token{kind: NUM_KW},
		token{kind: RPAREN},
		token{kind: FAT_ARR},
		token{kind: NUM_KW},
		token{kind: LBRACE},
		token{kind: RETURN_KW},
		literal{kind: IDENTIFIER, value: "n"},
		token{kind: RBRACE},
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
	if function.Operator.Kind() != FAT_ARR {
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
	parser := MakeParser(&testTokenizer{tokens: []Token{
		token{kind: LBRACKET},
		literal{kind: IDENTIFIER, value: "Type"},
		token{kind: RBRACKET},
		token{kind: LPAREN},
		token{kind: RPAREN},
		token{kind: FAT_ARR},
		token{kind: LBRACE},
		token{kind: RBRACE},
	}})
	node := ParseExpression(parser)

	_, ok := node.(FunctionExpression)
	if !ok {
		t.Fatalf("Expected FunctionExpression, got %#v", node)
		return
	}
}
