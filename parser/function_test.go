package parser

import "testing"

func TestSlimArrowFunctionWithoutArgs(t *testing.T) {
	parser := MakeParser(&testTokenizer{tokens: []Token{
		token{kind: LeftParenthesis},
		token{kind: RightParenthesis},
		token{kind: SlimArrow},
		literal{kind: NumberLiteral, value: "42"},
	}})
	node := parser.parseFunctionExpression()

	function, ok := node.(FunctionExpression)
	if !ok {
		t.Fatalf("Expected FunctionExpression, got %#v", node)
	}

	if function.Params.Expr != nil {
		t.Fatalf("Expected no params, got %#v", function.Params.Expr)
	}
	if function.Operator.Kind() != SlimArrow {
		t.Fatalf("Expected '->', got %v", function.Operator.Text())
	}
	if _, ok := function.Expr.(*Literal); !ok {
		t.Fatalf("Expected literal, got %#v", function.Expr)
	}
	if function.Body != nil {
		t.Fatalf("Expected no Body, got %#v", function.Body)
	}
}

func TestSlimArrowFunctionWithArgs(t *testing.T) {
	parser := MakeParser(&testTokenizer{tokens: []Token{
		token{kind: LeftParenthesis},
		literal{kind: Name, value: "n"},
		token{kind: NumberKeyword},
		token{kind: RightParenthesis},
		token{kind: SlimArrow},
		literal{kind: NumberLiteral, value: "2"},
		token{kind: Mul},
		literal{kind: Name, value: "n"},
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
	if function.Operator.Kind() != SlimArrow {
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
		token{kind: LeftParenthesis},
		token{kind: NumberKeyword},
		token{kind: RightParenthesis},
		token{kind: SlimArrow},
		token{kind: NumberKeyword},
	}})
	node := parser.parseFunctionExpression()

	function, ok := node.(FunctionExpression)
	if !ok {
		t.Fatalf("Expected FunctionExpression, got %#v", node)
	}

	if function.Operator.Kind() != SlimArrow {
		t.Fatalf("Expected '->', got %v", function.Operator.Text())
	}
	if _, ok := function.Expr.(*Literal); !ok {
		t.Fatalf("Expected literal, got %#v", function.Expr)
	}
	if function.Body != nil {
		t.Fatalf("Expected no Body, got %#v", function.Body)
	}
}

func TestFatArrowFunctionWithoutArgs(t *testing.T) {
	parser := MakeParser(&testTokenizer{tokens: []Token{
		token{kind: LeftParenthesis},
		token{kind: RightParenthesis},
		token{kind: FatArrow},
		token{kind: NumberKeyword},
		token{kind: LeftBrace},
		token{kind: ReturnKeyword},
		literal{kind: NumberLiteral, value: "42"},
		token{kind: RightBrace},
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
	if function.Operator.Kind() != FatArrow {
		t.Fatalf("Expected '=>', got %v", function.Operator.Text())
	}
	if _, ok := function.Expr.(*Literal); !ok {
		t.Fatalf("Expected literal, got %#v", function.Expr)
	}
	if function.Body == nil {
		t.Fatalf("Expected Body, got nothing")
	}
}

func TestFatArrowFunctionWithArgs(t *testing.T) {
	parser := MakeParser(&testTokenizer{tokens: []Token{
		token{kind: LeftParenthesis},
		literal{kind: Name, value: "n"},
		token{kind: NumberKeyword},
		token{kind: RightParenthesis},
		token{kind: FatArrow},
		token{kind: NumberKeyword},
		token{kind: LeftBrace},
		token{kind: ReturnKeyword},
		literal{kind: Name, value: "n"},
		token{kind: RightBrace},
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
	if function.Operator.Kind() != FatArrow {
		t.Fatalf("Expected '=>', got %v", function.Operator.Text())
	}
	if _, ok := function.Expr.(*Literal); !ok {
		t.Fatalf("Expected literal, got %#v", function.Expr)
	}
	if function.Body == nil {
		t.Fatalf("Expected Body, got nothing")
	}
}

func TestFunctionWithTypeArgs(t *testing.T) {
	parser := MakeParser(&testTokenizer{tokens: []Token{
		token{kind: LeftBracket},
		literal{kind: Name, value: "Type"},
		token{kind: RightBracket},
		token{kind: LeftParenthesis},
		token{kind: RightParenthesis},
		token{kind: FatArrow},
		token{kind: LeftBrace},
		token{kind: RightBrace},
	}})
	node := ParseExpression(parser)

	_, ok := node.(FunctionExpression)
	if !ok {
		t.Fatalf("Expected FunctionExpression, got %#v", node)
		return
	}
}
