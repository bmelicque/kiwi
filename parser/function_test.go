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

func TestFunctionExpressionWithArgs(t *testing.T) {
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

func TestCheckImplicitReturn(t *testing.T) {
	parser := MakeParser(nil)
	expr := &FunctionExpression{
		Params: &ParenthesizedExpression{},
		Body: &Block{Statements: []Node{
			&Literal{literal{kind: NumberLiteral, value: "42"}},
		}},
	}
	expr.typeCheck(parser)

	if len(parser.errors) > 0 {
		t.Fatalf("Expected no errors, got %#v", parser.errors)
	}
	if expr.returnType.Kind() != NUMBER {
		t.Fatalf("Expected number, got %v", expr.returnType.Kind())
	}
}

func TestCheckImplicitReturnBadReturns(t *testing.T) {
	parser := MakeParser(nil)
	parser.scope.Add("result", Loc{}, makeResultType(Nil{}, Number{}))
	// () => {
	//		if true {
	//			return false
	//		} else if true {
	//			throw false
	//		} else {
	//			try result
	//		}
	//		42
	// }
	expr := &FunctionExpression{
		Params: &ParenthesizedExpression{},
		Body: &Block{Statements: []Node{
			&IfExpression{
				Keyword:   token{kind: IfKeyword},
				Condition: &Literal{literal{kind: BooleanLiteral, value: "true"}},
				Body: &Block{Statements: []Node{
					&Exit{
						Operator: token{kind: ReturnKeyword},
						Value:    &Literal{literal{kind: BooleanLiteral, value: "false"}},
					},
				}},
				Alternate: &IfExpression{
					Keyword:   token{kind: IfKeyword},
					Condition: &Literal{literal{kind: BooleanLiteral, value: "true"}},
					Body: &Block{Statements: []Node{
						&Exit{
							Operator: token{kind: ThrowKeyword},
							Value:    &Literal{literal{kind: BooleanLiteral, value: "false"}},
						},
					}},
					Alternate: &Block{Statements: []Node{
						&TryExpression{
							Keyword: token{kind: TryKeyword},
							Expr:    &Identifier{Token: literal{kind: Name, value: "result"}},
						},
					}},
				},
			},
			&Literal{literal{kind: NumberLiteral, value: "42"}},
		}},
	}
	expr.typeCheck(parser)

	if len(parser.errors) != 5 {
		// expect 2 errors for if expression types
		// expect 1 error for early return
		// expect 1 error for try with implicit return type
		// expect 1 error for throw with implicit return type
		t.Fatalf("Expected 5 errors, got %v: %#v", len(parser.errors), parser.errors)
	}
}

func TestCheckExplicitReturn(t *testing.T) {
	parser := MakeParser(nil)
	expr := &FunctionExpression{
		Params: &ParenthesizedExpression{},
		Explicit: &BinaryExpression{
			Left:     &Literal{token{kind: StringKeyword}},
			Right:    &Literal{token{kind: NumberKeyword}},
			Operator: token{kind: Bang},
		},
		Body: &Block{Statements: []Node{
			&IfExpression{
				Keyword:   token{kind: IfKeyword},
				Condition: &Literal{literal{kind: BooleanLiteral}},
				Body: &Block{Statements: []Node{
					&Exit{
						Operator: token{kind: ReturnKeyword},
						Value:    &Literal{literal{kind: NumberLiteral, value: "42"}},
					},
				}},
				Alternate: &Block{Statements: []Node{
					&Exit{
						Operator: token{kind: ThrowKeyword},
						Value:    &Literal{literal{kind: StringLiteral, value: "\"\""}},
					},
				}},
			},
			&Literal{literal{kind: NumberLiteral, value: "42"}},
		}},
	}
	expr.typeCheck(parser)

	if len(parser.errors) != 1 {
		// expect 1 error for if expression types
		t.Fatalf("Expected 1 error, got %#v", parser.errors)
	}
	if alias, ok := expr.returnType.(TypeAlias); !ok || alias.Name != "Result" {
		t.Fatalf("Result type expected")
	}
}
