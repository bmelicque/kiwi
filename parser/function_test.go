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

func TestParseFunctionExpression(t *testing.T) {
	tests := []struct {
		name        string
		source      string
		wantError   bool
		expectedLoc Loc
	}{
		{
			name:        "no params, implicit return",
			source:      "() => {}",
			wantError:   false,
			expectedLoc: Loc{Position{1, 1}, Position{1, 9}},
		},
		{
			name:        "explicit return type",
			source:      "() => number {}",
			wantError:   false,
			expectedLoc: Loc{Position{1, 1}, Position{1, 16}},
		},
		{
			name:        "explicit void return",
			source:      "() => _ {}",
			wantError:   false,
			expectedLoc: Loc{Position{1, 1}, Position{1, 11}},
		},
		{
			name:        "one param",
			source:      "(n number) => {}",
			wantError:   false,
			expectedLoc: Loc{Position{1, 1}, Position{1, 17}},
		},
		{
			name:        "several params",
			source:      "(a number, b number) => {}",
			wantError:   false,
			expectedLoc: Loc{Position{1, 1}, Position{1, 27}},
		},
		{
			name:        "shortened params", // for HOF
			source:      "(a, b) => {}",
			wantError:   false,
			expectedLoc: Loc{Position{1, 1}, Position{1, 13}},
		},
		{
			name:        "type param",
			source:      "[Type]() => {}",
			wantError:   false,
			expectedLoc: Loc{Position{1, 1}, Position{1, 15}},
		},
		{
			name:        "missing body",
			source:      "() =>",
			wantError:   true,
			expectedLoc: Loc{Position{1, 1}, Position{1, 6}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := MakeParser(strings.NewReader(tt.source))
			expr := parser.parseBinaryExpression()
			if tt.wantError && len(parser.errors) == 0 {
				t.Error("Got no errors, want one\n")
			}
			if !tt.wantError && len(parser.errors) > 0 {
				t.Error("Got one error, want none\n")
			}
			if expr.Loc() != tt.expectedLoc {
				t.Errorf("Got loc %v, want %v", expr.Loc(), tt.expectedLoc)
			}
		})
	}
}

// ------------------------------------
// TEST TYPE-CHECK FUNCTION EXPRESSIONS
// ------------------------------------

func TestCheckFunctionExpression(t *testing.T) {
	tests := []struct {
		name         string
		expr         *FunctionExpression
		wantError    bool
		expectedType string
	}{
		{
			name: "no params, no return",
			expr: &FunctionExpression{
				Params: &ParenthesizedExpression{Expr: MakeTuple(nil)},
				Body:   MakeBlock([]Node{}),
			},
			wantError:    false,
			expectedType: "() -> ()",
		},
		{
			name: "no params, void return",
			expr: &FunctionExpression{
				Params:   &ParenthesizedExpression{Expr: MakeTuple(nil)},
				Explicit: &Identifier{Token: literal{kind: Name, value: "_"}},
				Body:     MakeBlock([]Node{}),
			},
			wantError:    false,
			expectedType: "() -> ()",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := MakeParser(strings.NewReader(""))
			tt.expr.typeCheck(parser)
			if tt.wantError && len(parser.errors) == 0 {
				t.Error("Got no errors, want one\n")
			}
			if !tt.wantError && len(parser.errors) > 0 {
				t.Error("Got one error, want none\n")
				t.Log(parser.errors[0].Text())
			}
			text := tt.expr.Type().Text()
			if text != tt.expectedType {
				t.Errorf("Got %v, want %v", text, tt.expectedType)
			}
		})
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

func TestTypeCheckReturnsExplicit(t *testing.T) {
	tests := []struct {
		name       string
		returnType ExpressionType
		body       *Block
		wantErr    int
	}{
		{
			name:       "matching return type",
			returnType: Type{String{}},
			body: &Block{
				Statements: []Node{
					&Exit{
						Operator: token{kind: ReturnKeyword},
						Value:    &Literal{literal{kind: StringLiteral}},
					},
				},
			},
			wantErr: 0,
		},
		{
			name:       "matching return type expecting result",
			returnType: Type{makeResultType(String{}, Number{})},
			body: &Block{
				Statements: []Node{
					&Exit{
						Operator: token{kind: ReturnKeyword},
						Value:    &Literal{literal{kind: StringLiteral}},
					},
				},
			},
			wantErr: 0,
		},
		{
			name:       "mismatched return type",
			returnType: Type{Number{}},
			body: &Block{
				Statements: []Node{
					&Exit{
						Operator: token{kind: ReturnKeyword},
						Value:    &Literal{literal{kind: StringLiteral}},
					},
				},
			},
			wantErr: 1,
		},
		{
			name:       "mismatched return type expecting result",
			returnType: Type{makeResultType(Number{}, Number{})},
			body: &Block{
				Statements: []Node{
					&Exit{
						Operator: token{kind: ReturnKeyword},
						Value:    &Literal{literal{kind: StringLiteral}},
					},
				},
			},
			wantErr: 1,
		},
		{
			name:       "multiple return statements",
			returnType: Type{String{}},
			body: &Block{
				Statements: []Node{
					&Exit{
						Operator: token{kind: ReturnKeyword},
						Value:    &Literal{literal{kind: StringLiteral}},
					},
					&Exit{
						Operator: token{kind: ReturnKeyword},
						Value:    &Literal{literal{kind: StringLiteral}},
					},
				},
			},
			wantErr: 0,
		},
		{
			name:       "body type",
			returnType: Type{String{}},
			body: &Block{
				Statements: []Node{&Literal{literal{kind: StringLiteral}}},
			},
			wantErr: 0,
		},
		{
			name:       "mismathced body type",
			returnType: Type{String{}},
			body: &Block{
				Statements: []Node{&Literal{literal{kind: NumberLiteral}}},
			},
			wantErr: 1,
		},
		{
			// error already handled in another function
			name:       "invalid (non-type) explicit type produce no error",
			returnType: Invalid{},
			body: &Block{
				Statements: []Node{&Literal{literal{kind: StringLiteral}}},
			},
			wantErr: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := MakeParser(strings.NewReader(""))

			typeCheckReturnsExplicit(p, tt.returnType, tt.body)

			if len(p.errors) != tt.wantErr {
				t.Errorf("expected %v errors but got %v", tt.wantErr, len(p.errors))
			}
		})
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
