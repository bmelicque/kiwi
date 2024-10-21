package parser

import "testing"

func TestFunctionType(t *testing.T) {
	// (number) -> number
	parser := MakeParser(&testTokenizer{tokens: []Token{
		token{kind: LeftParenthesis},
		token{kind: NumberKeyword},
		token{kind: RightParenthesis},
		token{kind: SlimArrow},
		token{kind: NumberKeyword},
	}})
	node := parser.parseFunctionExpression(nil)

	function, ok := node.(*FunctionTypeExpression)
	if !ok {
		t.Fatalf("Expected FunctionTypeExpression, got %#v", node)
	}

	if _, ok := function.Expr.(*Literal); !ok {
		t.Fatalf("Expected literal, got %#v", function.Expr)
	}
}

func TestFunctionExpressionWithoutArgs(t *testing.T) {
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
	node := parser.parseFunctionExpression(nil)

	function, ok := node.(*FunctionExpression)
	if !ok {
		t.Fatalf("Expected FunctionExpression, got %#v", node)
		return
	}

	params := function.Params.Expr.(*TupleExpression).Elements
	if len(params) > 0 {
		t.Fatalf("Expected no params, got %#v", function.Params.Expr)
	}
	if _, ok := function.Explicit.(*Literal); !ok {
		t.Fatalf("Expected literal, got %#v", function.Explicit)
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
	node := parser.parseFunctionExpression(nil)

	function, ok := node.(*FunctionExpression)
	if !ok {
		t.Fatalf("Expected FunctionExpression, got %#v", node)
		return
	}

	if len(function.Params.Expr.(*TupleExpression).Elements) != 1 {
		t.Fatalf("Expected 1 param, got %#v", function.Params.Expr)
	}
	if _, ok := function.Explicit.(*Literal); !ok {
		t.Fatalf("Expected literal, got %#v", function.Explicit)
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
	node := parser.parseExpression()

	_, ok := node.(*FunctionExpression)
	if !ok {
		t.Fatalf("Expected FunctionExpression, got %#v", node)
		return
	}
}
