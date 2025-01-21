package parser

import (
	"strings"
	"testing"
)

func TestParseUnaryExpression(t *testing.T) {
	type test struct {
		name        string
		source      string
		wantError   bool
		expectedLoc Loc
	}
	tests := []test{
		{
			name:        "option type",
			source:      "?number",
			wantError:   false,
			expectedLoc: Loc{Position{1, 1}, Position{1, 8}},
		},
		{
			name:        "result type",
			source:      "!number",
			wantError:   false,
			expectedLoc: Loc{Position{1, 1}, Position{1, 8}},
		},
		{
			name:        "try expression",
			source:      "try 42",
			wantError:   false,
			expectedLoc: Loc{Position{1, 1}, Position{1, 7}},
		},
		{
			name:        "async call",
			source:      "async fetch()",
			wantError:   false,
			expectedLoc: Loc{Position{1, 1}, Position{1, 14}},
		},
		{
			name:        "await expression",
			source:      "await promise",
			wantError:   false,
			expectedLoc: Loc{Position{1, 1}, Position{1, 14}},
		},
		{
			name:        "ref",
			source:      "&value",
			wantError:   false,
			expectedLoc: Loc{Position{1, 1}, Position{1, 7}},
		},
		{
			name:        "deref",
			source:      "*ref",
			wantError:   false,
			expectedLoc: Loc{Position{1, 1}, Position{1, 5}},
		},
		{
			name:        "nested unary expressions",
			source:      "??number",
			wantError:   false,
			expectedLoc: Loc{Position{1, 1}, Position{1, 9}},
		},
		{
			name:        "missing operand produce one error",
			source:      "?",
			wantError:   true,
			expectedLoc: Loc{Position{1, 1}, Position{1, 2}},
		},
		{
			name:        "awaiting on non-call produces one error",
			source:      "async true",
			wantError:   true,
			expectedLoc: Loc{Position{1, 1}, Position{1, 11}},
		},
		{
			name:        "list type",
			source:      "[]number",
			wantError:   false,
			expectedLoc: Loc{Position{1, 1}, Position{1, 9}},
		},
		{
			name:        "empty list type",
			source:      "[]",
			wantError:   true,
			expectedLoc: Loc{Position{1, 1}, Position{1, 3}},
		},
		{
			name:        "list type with something in brackets",
			source:      "[number]number",
			wantError:   true,
			expectedLoc: Loc{Position{1, 1}, Position{1, 15}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := MakeParser(strings.NewReader(tt.source))
			expr := parser.parseUnaryExpression()
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

func TestCheckUnaryExpression(t *testing.T) {
	type test struct {
		name         string
		expr         Expression
		wantError    bool
		expectedType string
	}
	tests := []test{
		{
			name: "async non call",
			expr: &UnaryExpression{
				Operator: token{kind: AsyncKeyword},
				Operand:  &Identifier{Token: literal{kind: Name, value: "value"}},
			},
			wantError:    false, // error handled in parsing
			expectedType: "async[invalid]",
		},
		{
			name: "async call of asyncable function",
			expr: &UnaryExpression{
				Operator: token{kind: AsyncKeyword},
				Operand: &CallExpression{
					Callee: &Identifier{Token: literal{kind: Name, value: "asyncGetter"}},
					Args:   &ParenthesizedExpression{Expr: &TupleExpression{Elements: []Expression{}}},
				},
			},
			wantError:    false,
			expectedType: "async[number]",
		},
		{
			name: "async call of non-asyncable function",
			expr: &UnaryExpression{
				Operator: token{kind: AsyncKeyword},
				Operand: &CallExpression{
					Callee: &Identifier{Token: literal{kind: Name, value: "getter"}},
					Args:   &ParenthesizedExpression{Expr: &TupleExpression{Elements: []Expression{}}},
				},
			},
			wantError:    true,
			expectedType: "async[number]",
		},
		{
			name: "await promise",
			expr: &UnaryExpression{
				Operator: token{kind: AwaitKeyword},
				Operand:  &Identifier{Token: literal{kind: Name, value: "request"}},
			},
			wantError:    false,
			expectedType: "number",
		},
		{
			name: "await non-promise",
			expr: &UnaryExpression{
				Operator: token{kind: AwaitKeyword},
				Operand:  &Identifier{Token: literal{kind: Name, value: "value"}},
			},
			wantError:    true,
			expectedType: "number",
		},
		{
			name: "logical not",
			expr: &UnaryExpression{
				Operator: token{kind: Bang},
				Operand:  &Literal{literal{kind: BooleanLiteral}},
			},
			wantError:    false,
			expectedType: "boolean",
		},
		{
			name: "error type",
			expr: &UnaryExpression{
				Operator: token{kind: Bang},
				Operand:  &Literal{literal{kind: NumberKeyword}},
			},
			wantError:    false,
			expectedType: "(!number)",
		},
		{
			name: "bang on (not type, not bool)",
			expr: &UnaryExpression{
				Operator: token{kind: Bang},
				Operand:  &Literal{literal{kind: StringLiteral}},
			},
			wantError:    true,
			expectedType: "invalid",
		},
		{
			name: "ref of value",
			expr: &UnaryExpression{
				Operator: token{kind: BinaryAnd},
				Operand:  &Identifier{Token: literal{kind: Name, value: "value"}},
			},
			wantError:    false,
			expectedType: "&number",
		},
		{
			name: "ref of type",
			expr: &UnaryExpression{
				Operator: token{kind: BinaryAnd},
				Operand:  &Identifier{Token: literal{kind: Name, value: "Type"}},
			},
			wantError:    false,
			expectedType: "(&Type)",
		},
		{
			name: "deref of ref",
			expr: &UnaryExpression{
				Operator: token{kind: Mul},
				Operand:  &Identifier{Token: literal{kind: Name, value: "ref"}},
			},
			wantError:    false,
			expectedType: "number",
		},
		{
			name: "deref of non-ref",
			expr: &UnaryExpression{
				Operator: token{kind: Mul},
				Operand:  &Identifier{Token: literal{kind: Name, value: "value"}},
			},
			wantError:    true,
			expectedType: "invalid",
		},
		{
			name: "option type",
			expr: &UnaryExpression{
				Operator: token{kind: QuestionMark},
				Operand:  &Identifier{Token: literal{kind: Name, value: "Type"}},
			},
			wantError:    false,
			expectedType: "(?Type)",
		},
		{
			name: "option values are invalid",
			expr: &UnaryExpression{
				Operator: token{kind: QuestionMark},
				Operand:  &Identifier{Token: literal{kind: Name, value: "value"}},
			},
			wantError:    true,
			expectedType: "(?invalid)",
		},
		{
			name: "try result",
			expr: &UnaryExpression{
				Operator: token{kind: TryKeyword},
				Operand:  &Identifier{Token: literal{kind: Name, value: "result"}},
			},
			wantError:    false,
			expectedType: "number",
		},
		{
			name: "try simple type",
			expr: &UnaryExpression{
				Operator: token{kind: TryKeyword},
				Operand:  &Identifier{Token: literal{kind: Name, value: "value"}},
			},
			wantError:    true,
			expectedType: "invalid",
		},
		{
			name: "list type",
			expr: &ListTypeExpression{
				Bracketed: &BracketedExpression{},
				Expr:      &Identifier{Token: literal{kind: Name, value: "Type"}},
			},
			wantError:    false,
			expectedType: "([]Type)",
		},
		{
			name: "list type without operand",
			expr: &ListTypeExpression{
				Bracketed: &BracketedExpression{},
				Expr:      nil,
			},
			wantError:    false, // error already handled in parsing
			expectedType: "([]invalid)",
		},
		{
			name: "list type with value operand",
			expr: &ListTypeExpression{
				Bracketed: &BracketedExpression{},
				Expr:      &Identifier{Token: literal{kind: Name, value: "value"}},
			},
			wantError:    true,
			expectedType: "([]invalid)",
		},
		{
			name: "list type with something inside brackets",
			expr: &ListTypeExpression{
				Bracketed: &BracketedExpression{Expr: &Literal{token{kind: NumberKeyword}}},
				Expr:      &Identifier{Token: literal{kind: Name, value: "Type"}},
			},
			wantError:    false, // handled in parsing
			expectedType: "([]Type)",
		},
	}

	scope := NewScope(ProgramScope)
	scope.Add("value", Loc{}, Number{})
	scope.Add("ref", Loc{}, Ref{Number{}})
	scope.Add("Type", Loc{}, Type{TypeAlias{Name: "Type"}})
	scope.Add("result", Loc{}, makeResultType(Number{}, String{}))
	scope.Add("getter", Loc{}, newGetter(Number{}))
	scope.Add("asyncGetter", Loc{}, Function{Params: &Tuple{}, Returned: Number{}, Async: true})
	scope.Add("request", Loc{}, makePromise(Number{}))

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := MakeParser(strings.NewReader(""))
			parser.scope = scope
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
				t.Errorf("expected %v, got %v", tt.expectedType, text)
			}
		})
	}
}

func TestParseReferenceBadOperand(t *testing.T) {
	parser := MakeParser(strings.NewReader("&value()"))
	parser.parseExpression()

	if len(parser.errors) != 1 {
		t.Fatalf("Expected 1 errors, got %+v: %#v", len(parser.errors), parser.errors)
	}
}
