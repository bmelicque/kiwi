package parser

import (
	"strings"
	"testing"
)

func TestParseBinaryExpression(t *testing.T) {
	tests := []struct {
		name             string
		source           string
		wantError        bool
		expectedOperator TokenKind
		expectedLoc      Loc
	}{
		{
			name:             "valid map",
			source:           "Key#Value",
			wantError:        false,
			expectedOperator: Hash,
			expectedLoc:      Loc{Position{1, 1}, Position{1, 10}},
		},
		{
			name:             "map missing lhs",
			source:           "#Value",
			wantError:        true,
			expectedOperator: Hash,
			expectedLoc:      Loc{Position{1, 1}, Position{1, 7}},
		},
		{
			name:             "map missing rhs",
			source:           "Key#",
			wantError:        true,
			expectedOperator: Hash,
			expectedLoc:      Loc{Position{1, 1}, Position{1, 5}},
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
			if expr.(*BinaryExpression).Operator.Kind() != tt.expectedOperator {
				t.Errorf("Bad operator")
			}
			if expr.Loc() != tt.expectedLoc {
				t.Errorf("Got loc %v, want %v", expr.Loc(), tt.expectedLoc)
			}
		})
	}
}

func TestCheckBinaryExpression(t *testing.T) {
	tests := []struct {
		name         string
		expr         *BinaryExpression
		wantError    bool
		expectedType string
	}{
		{
			name: "valid map",
			expr: &BinaryExpression{
				Left:     &Literal{token{kind: StringKeyword}},
				Operator: token{kind: Hash},
				Right:    &Literal{token{kind: NumberKeyword}},
			},
			wantError:    false,
			expectedType: "(string#number)",
		},
		{
			name: "map without lhs",
			expr: &BinaryExpression{
				Left:     nil,
				Operator: token{kind: Hash},
				Right:    &Literal{token{kind: NumberKeyword}},
			},
			wantError:    false, // handled in parser
			expectedType: "(invalid#number)",
		},
		{
			name: "map without rhs",
			expr: &BinaryExpression{
				Left:     &Literal{token{kind: StringKeyword}},
				Operator: token{kind: Hash},
				Right:    nil,
			},
			wantError:    false, // handled in parser,
			expectedType: "(string#invalid)",
		},
		{
			name: "map with non-type on lhs",
			expr: &BinaryExpression{
				Left:     &Literal{token{kind: StringLiteral}},
				Operator: token{kind: Hash},
				Right:    &Literal{token{kind: NumberKeyword}},
			},
			wantError:    true,
			expectedType: "(invalid#number)",
		},
		{
			name: "map with non-type on rhs",
			expr: &BinaryExpression{
				Left:     &Literal{token{kind: StringKeyword}},
				Operator: token{kind: Hash},
				Right:    &Literal{token{kind: NumberLiteral}},
			},
			wantError:    true,
			expectedType: "(string#invalid)",
		},
		{
			name: "valid error type",
			expr: &BinaryExpression{
				Left:     &Literal{token{kind: StringKeyword}},
				Operator: token{kind: Bang},
				Right:    &Literal{token{kind: NumberKeyword}},
			},
			wantError:    false,
			expectedType: "(string!number)",
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
			}
			text := tt.expr.Type().Text()
			if text != tt.expectedType {
				t.Errorf("Got %v, want %v", text, tt.expectedType)
			}
		})
	}
}

func TestBinaryExpression(t *testing.T) {
	parser := MakeParser(strings.NewReader("2 ** 3"))
	expr := parser.parseBinaryExpression()

	if len(parser.errors) > 0 {
		t.Fatalf("Expected no parsing errors, got %#v", parser.errors)
	}

	binary, ok := expr.(*BinaryExpression)
	if !ok {
		t.Fatalf("Expected a binary expression, got %#v", expr)
	}
	_ = binary
}

func TestBinaryErrorType(t *testing.T) {
	parser := MakeParser(strings.NewReader("ErrType!OkType"))
	expr := parser.parseExpression()

	if len(parser.errors) > 0 {
		t.Fatalf("Expected no parsing errors, got %#v", parser.errors)
	}

	binary, ok := expr.(*BinaryExpression)
	if !ok {
		t.Fatalf("Expected a binary expression, got %#v", expr)
	}
	_ = binary
}

func TestCheckArithmeticExpression(t *testing.T) {
	parser := MakeParser(nil)
	expr := &BinaryExpression{
		Left:  &Literal{literal{kind: NumberLiteral}},
		Right: &Literal{literal{kind: NumberLiteral}},
	}

	operators := []TokenKind{Add, Sub, Mul, Div, Pow, Greater, GreaterEqual, Less, LessEqual}
	for _, operator := range operators {
		expr.Operator = token{kind: operator}
		expr.typeCheck(parser)
	}

	if len(parser.errors) > 0 {
		t.Fatalf("Expected no parsing errors, got %#v", parser.errors)
	}

	expr = &BinaryExpression{
		Left:  &Literal{literal{kind: BooleanLiteral}},
		Right: &Literal{literal{kind: BooleanLiteral}},
	}
	for _, operator := range operators {
		expr.Operator = token{kind: operator}
		expr.typeCheck(parser)
	}
	if len(parser.errors) != 2*len(operators) {
		t.Fatalf(
			"Expected %v parsing errors, got %#v",
			2*len(operators),
			parser.errors,
		)
	}
}

func TestCheckLogicalExpression(t *testing.T) {
	parser := MakeParser(nil)
	expr := &BinaryExpression{
		Left:  &Literal{literal{kind: BooleanLiteral}},
		Right: &Literal{literal{kind: BooleanLiteral}},
	}

	operators := []TokenKind{LogicalAnd, LogicalOr}
	for _, operator := range operators {
		expr.Operator = token{kind: operator}
		expr.typeCheck(parser)
	}

	if len(parser.errors) > 0 {
		t.Fatalf("Expected no parsing errors, got %#v", parser.errors)
	}

	expr = &BinaryExpression{
		Left:  &Literal{literal{kind: StringLiteral}},
		Right: &Literal{literal{kind: StringLiteral}},
	}
	for _, operator := range operators {
		expr.Operator = token{kind: operator}
		expr.typeCheck(parser)
	}
	if len(parser.errors) != 2*len(operators) {
		t.Fatalf(
			"Expected %v parsing errors, got %#v",
			2*len(operators),
			parser.errors,
		)
	}
}

func TestCheckConcatExpression(t *testing.T) {
	parser := MakeParser(nil)
	expr := &BinaryExpression{
		Left:     &Literal{literal{kind: StringLiteral}},
		Right:    &Literal{literal{kind: StringLiteral}},
		Operator: token{kind: Concat},
	}
	expr.typeCheck(parser)

	if len(parser.errors) > 0 {
		t.Fatalf("Expected no parsing errors, got %#v", parser.errors)
	}

	parser.scope.Add("a", Loc{}, List{Number{}})
	parser.scope.Add("b", Loc{}, List{Number{}})
	parser.scope.Add("c", Loc{}, List{String{}})
	expr = &BinaryExpression{
		Left:     &Identifier{Token: literal{kind: Name, value: "a"}},
		Right:    &Identifier{Token: literal{kind: Name, value: "b"}},
		Operator: token{kind: Concat},
	}
	expr.typeCheck(parser)
	if len(parser.errors) > 0 {
		t.Fatalf("Expected no parsing errors, got %#v", parser.errors)
	}

	expr = &BinaryExpression{
		Left:     &Identifier{Token: literal{kind: Name, value: "a"}},
		Right:    &Identifier{Token: literal{kind: Name, value: "c"}},
		Operator: token{kind: Concat},
	}
	expr.typeCheck(parser)
	if len(parser.errors) != 1 {
		t.Fatalf("Expected 1 parsing errors, got %#v", parser.errors)
	}
}

func TestCheckComparisonExpression(t *testing.T) {
	parser := MakeParser(nil)
	expr := &BinaryExpression{
		Left:  &Literal{literal{kind: BooleanLiteral}},
		Right: &Literal{literal{kind: BooleanLiteral}},
	}

	operators := []TokenKind{Equal, NotEqual}
	for _, operator := range operators {
		expr.Operator = token{kind: operator}
		expr.typeCheck(parser)
	}

	if len(parser.errors) > 0 {
		t.Fatalf("Expected no parsing errors, got %#v", parser.errors)
	}

	expr = &BinaryExpression{
		Left:  &Literal{literal{kind: BooleanLiteral}},
		Right: &Literal{literal{kind: StringLiteral}},
	}
	for _, operator := range operators {
		expr.Operator = token{kind: operator}
		expr.typeCheck(parser)
	}
	if len(parser.errors) != len(operators) {
		t.Fatalf(
			"Expected %v parsing errors, got %#v",
			len(operators),
			parser.errors,
		)
	}
}
