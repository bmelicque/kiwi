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
	loc := Loc{Position{1, 1}, Position{1, 9}}
	if node.Loc() != loc {
		t.Fatalf("Expected loc %v, got %v", loc, node.Loc())
	}

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

func TestParseHOFFunctionArgument(t *testing.T) {
	parser := MakeParser(strings.NewReader("(a, b number) => {}"))
	parser.parseFunctionExpression(nil)
	testParserErrors(t, parser, 0)
}

func TestParseFunctionExpressionMissingBody(t *testing.T) {
	parser := MakeParser(strings.NewReader("() =>"))
	expr := parser.parseFunctionExpression(nil)
	testParserErrors(t, parser, 1)
	loc := Loc{Position{1, 1}, Position{1, 6}}
	if expr.Loc() != loc {
		t.Fatalf("Expected loc %v, got %v", loc, expr.Loc())
	}
}

func TestParseFunctionExpressionExplicit(t *testing.T) {
	parser := MakeParser(strings.NewReader("() => number {}"))
	node := parser.parseFunctionExpression(nil)
	testParserErrors(t, parser, 0)

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

func TestParseFunctionWithTypeParams(t *testing.T) {
	parser := MakeParser(strings.NewReader("[Type]() => {}"))
	node := parser.parseExpression()
	testParserErrors(t, parser, 0)
	loc := Loc{Position{1, 1}, Position{1, 15}}
	if node.Loc() != loc {
		t.Fatalf("Expected loc %v, got %v", loc, node.Loc())
	}

	f, ok := node.(*FunctionExpression)
	if !ok {
		t.Fatalf("Expected FunctionExpression, got %#v", node)
	}
	if f.TypeParams == nil {
		t.Fatalf("Expected type params, got nothing")
	}
}

// ------------------------------------
// TEST TYPE-CHECK FUNCTION EXPRESSIONS
// ------------------------------------

func TestCheckFunctionExpressionNoParams(t *testing.T) {
	parser := MakeParser(nil)
	expr := &FunctionExpression{
		Params: &ParenthesizedExpression{Expr: &TupleExpression{Elements: []Expression{}}},
		Body:   &Block{Statements: []Node{}},
	}
	expr.typeCheck(parser)
	testParserErrors(t, parser, 0)

	params := expr.Type().(Function).Params.Elements
	if len(params) != 0 {
		t.Fatalf("Expected no params, found %#v", params)
	}
}

func TestCheckFunctionExpressionParams(t *testing.T) {
	parser := MakeParser(nil)
	expr := &FunctionExpression{
		Params: &ParenthesizedExpression{Expr: &TupleExpression{Elements: []Expression{
			&Param{
				Identifier: &Identifier{Token: literal{kind: Name, value: "a"}},
				Complement: &Literal{token{kind: NumberKeyword}},
			},
			&Param{
				Identifier: &Identifier{Token: literal{kind: Name, value: "b"}},
				Complement: &Literal{token{kind: NumberKeyword}},
			},
		}}},
		Body: &Block{Statements: []Node{
			&BinaryExpression{
				Left:     &Identifier{Token: literal{kind: Name, value: "a"}},
				Right:    &Identifier{Token: literal{kind: Name, value: "b"}},
				Operator: token{kind: Add},
			},
		}},
	}
	expr.typeCheck(parser)
	testParserErrors(t, parser, 0)

	if a, ok := expr.Body.scope.Find("a"); !ok || a.Typing.Text() != "number" {
		t.Log("Cannot find name 'a'")
		t.Fail()
	}
	if b, ok := expr.Body.scope.Find("b"); !ok || b.Typing.Text() != "number" {
		t.Log("Cannot find name 'b'")
		t.Fail()
	}
}

func TestCheckFunctionExpressionBadParam(t *testing.T) {
	parser := MakeParser(nil)
	expr := &FunctionExpression{
		Params: &ParenthesizedExpression{Expr: &TupleExpression{Elements: []Expression{
			&Param{
				Identifier: &Identifier{Token: literal{kind: Name, value: "a"}},
				Complement: &Literal{literal{kind: NumberLiteral, value: "42"}},
			},
		}}},
		Body: &Block{Statements: []Node{
			&Identifier{Token: literal{kind: Name, value: "a"}},
		}},
	}
	expr.typeCheck(parser)
	testParserErrors(t, parser, 1)

	if a, ok := expr.Body.scope.Find("a"); !ok || a.Typing != (Invalid{}) {
		t.Log("Expected 'a' to be unknown")
		t.Fail()
	}
}

func TestCheckImplicitReturn(t *testing.T) {
	parser := MakeParser(nil)
	expr := &FunctionExpression{
		Params: &ParenthesizedExpression{Expr: &TupleExpression{}},
		Body: &Block{Statements: []Node{
			&Literal{literal{kind: NumberLiteral, value: "42"}},
		}},
	}
	expr.typeCheck(parser)
	testParserErrors(t, parser, 0)
	if _, ok := expr.typing.Returned.(Number); !ok {
		t.Fatalf("Expected number, got %v", expr)
	}
}

func TestCheckImplicitReturnBadReturns(t *testing.T) {
	parser := MakeParser(nil)
	parser.scope.Add("result", Loc{}, makeResultType(Void{}, Number{}))
	expr := &FunctionExpression{
		Params: &ParenthesizedExpression{Expr: &TupleExpression{}},
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
			},
			&IfExpression{
				Keyword:   token{kind: IfKeyword},
				Condition: &Literal{literal{kind: BooleanLiteral, value: "true"}},
				Body: &Block{Statements: []Node{
					&Exit{
						Operator: token{kind: ThrowKeyword},
						Value:    &Literal{literal{kind: BooleanLiteral, value: "false"}},
					},
				}},
			},
			&UnaryExpression{
				Operator: token{kind: TryKeyword},
				Operand:  &Identifier{Token: literal{kind: Name, value: "result"}},
			},
		}},
	}
	expr.typeCheck(parser)

	// expect 1 error for early return
	// expect 1 error for try with implicit return type
	// expect 1 error for throw with implicit return type
	testParserErrors(t, parser, 3)
}

func TestCheckExplicitReturn(t *testing.T) {
	parser := MakeParser(nil)
	expr := &FunctionExpression{
		Params:   &ParenthesizedExpression{Expr: &TupleExpression{}},
		Explicit: &Literal{token{kind: NumberKeyword}},
		Body: &Block{Statements: []Node{
			&Literal{literal{kind: NumberLiteral, value: "42"}},
		}},
	}
	expr.typeCheck(parser)
	testParserErrors(t, parser, 0)

	if _, ok := expr.typing.Returned.(Number); !ok {
		t.Fatalf("Number type expected")
	}
}

func TestCheckBadExplicit(t *testing.T) {
	parser := MakeParser(nil)
	expr := &FunctionExpression{
		Params:   &ParenthesizedExpression{Expr: &TupleExpression{}},
		Explicit: &Literal{literal{kind: NumberLiteral, value: "42"}},
		Body: &Block{Statements: []Node{
			&Literal{literal{kind: NumberLiteral, value: "42"}},
		}},
	}
	expr.typeCheck(parser)
	testParserErrors(t, parser, 1)

	if expr.typing.Returned != (Invalid{}) {
		t.Fatalf("unknown expected")
	}
}

func TestCheckExplicitBadLastExpr(t *testing.T) {
	parser := MakeParser(nil)
	expr := &FunctionExpression{
		Params:   &ParenthesizedExpression{Expr: &TupleExpression{}},
		Explicit: &Literal{token{kind: NumberKeyword}},
		Body: &Block{Statements: []Node{
			&Literal{literal{kind: BooleanLiteral, value: "true"}},
		}},
	}
	expr.typeCheck(parser)
	testParserErrors(t, parser, 1)

	if _, ok := expr.typing.Returned.(Number); !ok {
		t.Fatalf("Number type expected")
	}
}

func TestCheckExplicitResult(t *testing.T) {
	parser := MakeParser(nil)
	expr := &FunctionExpression{
		Params: &ParenthesizedExpression{Expr: &TupleExpression{}},
		Explicit: &BinaryExpression{
			Left:     &Literal{token{kind: StringKeyword}},
			Right:    &Literal{token{kind: NumberKeyword}},
			Operator: token{kind: Bang},
		},
		Body: &Block{Statements: []Node{
			&Literal{literal{kind: NumberLiteral, value: "42"}},
		}},
	}
	expr.typeCheck(parser)
	testParserErrors(t, parser, 0)

	if alias, ok := expr.typing.Returned.(TypeAlias); !ok || alias.Name != "!" {
		t.Fatalf("Result type expected")
	}
}

func TestCheckExplicitThrow(t *testing.T) {
	parser := MakeParser(nil)
	expr := &FunctionExpression{
		Params: &ParenthesizedExpression{Expr: &TupleExpression{}},
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
						Operator: token{kind: ThrowKeyword},
						Value:    &Literal{literal{kind: StringLiteral, value: "\"\""}},
					},
				}},
			},
			&Literal{literal{kind: NumberLiteral, value: "42"}},
		}},
	}
	expr.typeCheck(parser)
	testParserErrors(t, parser, 0)

	if alias, ok := expr.typing.Returned.(TypeAlias); !ok || alias.Name != "!" {
		t.Fatalf("Result type expected")
	}
}

func TestCheckExplicitBadThrow(t *testing.T) {
	parser := MakeParser(nil)
	expr := &FunctionExpression{
		Params: &ParenthesizedExpression{Expr: &TupleExpression{}},
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
						Operator: token{kind: ThrowKeyword},
						Value:    &Literal{literal{kind: NumberLiteral, value: "42"}},
					},
				}},
			},
			&Literal{literal{kind: NumberLiteral, value: "42"}},
		}},
	}
	expr.typeCheck(parser)
	testParserErrors(t, parser, 1)
}

func TestCheckExplicitBadReturn(t *testing.T) {
	parser := MakeParser(nil)
	expr := &FunctionExpression{
		Params: &ParenthesizedExpression{Expr: &TupleExpression{}},
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
						Value:    &Literal{literal{kind: BooleanLiteral, value: "true"}},
					},
				}},
			},
			&Literal{literal{kind: NumberLiteral, value: "42"}},
		}},
	}
	expr.typeCheck(parser)
	testParserErrors(t, parser, 1)
}

func TestCheckAsyncFunctionExpression(t *testing.T) {
	parser := MakeParser(nil)
	parser.scope.Add("fetch", Loc{}, Function{
		Params:   &Tuple{},
		Returned: String{},
		Async:    true,
	})
	expr := &FunctionExpression{
		Params: &ParenthesizedExpression{Expr: &TupleExpression{}},
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
