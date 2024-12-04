package parser

import (
	"strings"
	"testing"
)

func TestFunctionType(t *testing.T) {
	parser := MakeParser(strings.NewReader("(number) -> number"))
	node := parser.parseFunctionExpression(nil)

	function, ok := node.(*FunctionTypeExpression)
	if !ok {
		t.Fatalf("Expected FunctionTypeExpression, got %#v", node)
	}

	if _, ok := function.Expr.(*Literal); !ok {
		t.Fatalf("Expected literal, got %#v", function.Expr)
	}
}

// ---------------------------------
// TEST PARSING FUNCTION EXPRESSIONS
// ---------------------------------

func TestParseFunctionExpressionWithNoParams(t *testing.T) {
	parser := MakeParser(strings.NewReader("() => {}"))
	node := parser.parseFunctionExpression(nil)
	testParserErrors(t, parser, 0)

	function, ok := node.(*FunctionExpression)
	if !ok {
		t.Fatalf("Expected FunctionExpression, got %#v", node)
	}

	params := function.Params.Expr.(*TupleExpression).Elements
	if len(params) > 0 {
		t.Fatalf("Expected no params, got %#v", function.Params.Expr)
	}
	if function.Body == nil {
		t.Fatalf("Expected Body, got nothing")
	}
}

func TestParseFunctionExpressionWithOneParam(t *testing.T) {
	parser := MakeParser(strings.NewReader("(n number) => {}"))
	node := parser.parseFunctionExpression(nil)
	testParserErrors(t, parser, 0)

	function, ok := node.(*FunctionExpression)
	if !ok {
		t.Fatalf("Expected FunctionExpression, got %#v", node)
	}

	params := function.Params.Expr.(*TupleExpression).Elements
	if len(params) != 1 {
		t.Fatalf("Expected 1 param, got %#v", function.Params.Expr)
	}
	if function.Body == nil {
		t.Fatalf("Expected Body, got nothing")
	}
}

func TestParseFunctionExpressionWithSeveralParams(t *testing.T) {
	parser := MakeParser(strings.NewReader("(a number, b number) => {}"))
	node := parser.parseFunctionExpression(nil)
	testParserErrors(t, parser, 0)

	function, ok := node.(*FunctionExpression)
	if !ok {
		t.Fatalf("Expected FunctionExpression, got %#v", node)
	}

	params := function.Params.Expr.(*TupleExpression).Elements
	if len(params) != 2 {
		t.Fatalf("Expected 2 params, got %#v", function.Params.Expr)
	}
	if function.Body == nil {
		t.Fatalf("Expected Body, got nothing")
	}
}

func TestParseFunctionExpressionWithBadParamNode(t *testing.T) {
	parser := MakeParser(strings.NewReader("(a, b number) => {}"))
	parser.parseFunctionExpression(nil)
	testParserErrors(t, parser, 1)
}

func TestParseFunctionExpressionExplicit(t *testing.T) {
	parser := MakeParser(strings.NewReader("() => number {}"))
	node := parser.parseFunctionExpression(nil)

	function, ok := node.(*FunctionExpression)
	if !ok {
		t.Fatalf("Expected FunctionExpression, got %#v", node)
		return
	}

	if _, ok := function.Explicit.(*Literal); !ok {
		t.Fatalf("Expected literal, got %#v", function.Explicit)
	}
	if function.Body == nil {
		t.Fatalf("Expected Body, got nothing")
	}
}

func TestParseFunctionWithTypeArgs(t *testing.T) {
	parser := MakeParser(strings.NewReader("[Type]() => {}"))
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
	if _, ok := expr.returnType.(Number); !ok {
		t.Fatalf("Expected number, got %v", expr)
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
						&UnaryExpression{
							Operator: token{kind: TryKeyword},
							Operand:  &Identifier{Token: literal{kind: Name, value: "result"}},
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
	if alias, ok := expr.returnType.(TypeAlias); !ok || alias.Name != "!" {
		t.Fatalf("Result type expected")
	}
}

func TestCheckAsync(t *testing.T) {
	parser := MakeParser(nil)
	parser.scope.Add("fetch", Loc{}, Function{
		Params:   &Tuple{},
		Returned: String{},
		Async:    true,
	})
	expr := &FunctionExpression{
		Params: &ParenthesizedExpression{},
		Body: &Block{Statements: []Node{
			&CallExpression{
				Callee: &Identifier{Token: literal{kind: Name, value: "fetch"}},
				Args:   &ParenthesizedExpression{Expr: &TupleExpression{}},
			},
		}},
	}
	expr.typeCheck(parser)

	if len(parser.errors) > 0 {
		t.Fatalf("Expected no errors, got %#v", parser.errors)
	}
	if !expr.canBeAsync {
		t.Fatalf("Expected function to be async, got %#v", parser.errors)
	}
}
