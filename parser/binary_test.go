package parser

import (
	"strings"
	"testing"
)

func TestBinaryExpression(t *testing.T) {
	parser, err := MakeParser(strings.NewReader("2 ** 3"))
	if err != nil {
		t.Fatal(err)
	}
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
	parser, err := MakeParser(strings.NewReader("ErrType!OkType"))
	if err != nil {
		t.Fatal(err)
	}
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
	parser, _ := MakeParser(nil)
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
	parser, _ := MakeParser(nil)
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
	parser, _ := MakeParser(nil)
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
	parser, _ := MakeParser(nil)
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

func TestCheckBinaryErrorType(t *testing.T) {
	parser, _ := MakeParser(nil)
	expr := &BinaryExpression{
		Left:     &Literal{literal{kind: StringKeyword}},
		Right:    &Literal{literal{kind: NumberKeyword}},
		Operator: token{kind: Bang},
	}
	expr.typeCheck(parser)

	if len(parser.errors) > 0 {
		t.Fatalf("Expected no parsing errors, got %#v", parser.errors)
	}
	ty, ok := expr.Type().(Type)
	if !ok {
		t.Fatalf("Type expected")
	}
	alias, ok := ty.Value.(TypeAlias)
	if !ok || alias.Name != "!" {
		t.Fatalf("Result type expected")
	}
	okType := alias.Ref.(Sum).getMember("Ok")
	if _, ok := okType.(Number); !ok {
		t.Fatalf("Number expected")
	}
}
