package parser

import (
	"strings"
	"testing"
)

func TestParseAsyncExpression(t *testing.T) {
	parser := MakeParser(strings.NewReader("async fetch()"))
	parser.parseUnaryExpression()

	if len(parser.errors) > 0 {
		t.Fatalf("Expected no errors, got %#v", parser.errors)
	}
}

func TestParseAsyncNoExpr(t *testing.T) {
	parser := MakeParser(strings.NewReader("async"))
	parser.parseUnaryExpression()

	if len(parser.errors) != 1 {
		t.Fatalf("Expected 1 error, got %#v", parser.errors)
	}
}

func TestParseAsyncNotCall(t *testing.T) {
	parser := MakeParser(strings.NewReader("async fetch"))
	parser.parseUnaryExpression()

	if len(parser.errors) != 1 {
		t.Fatalf("Expected 1 error, got %#v", parser.errors)
	}
}

func TestCheckAsyncExpression(t *testing.T) {
	parser := MakeParser(nil)
	parser.scope.Add("fetch", Loc{}, Function{
		Params:   &Tuple{},
		Returned: String{},
		Async:    true,
	})
	expr := &UnaryExpression{
		Operator: token{kind: AsyncKeyword},
		Operand: &CallExpression{
			Callee: &Identifier{Token: literal{kind: Name, value: "fetch"}},
			Args:   &ParenthesizedExpression{Expr: &TupleExpression{}},
		},
	}
	expr.typeCheck(parser)
	if len(parser.errors) != 0 {
		t.Fatalf("Expected no errors, got %#v", parser.errors)
	}
}

func TestParseAwaitExpression(t *testing.T) {
	parser := MakeParser(strings.NewReader("await request"))
	parser.parseUnaryExpression()

	if len(parser.errors) > 0 {
		t.Fatalf("Expected no errors, got %#v", parser.errors)
	}
}

func TestParseAwaitExpressionNoExpr(t *testing.T) {
	parser := MakeParser(strings.NewReader("await"))
	parser.parseUnaryExpression()

	if len(parser.errors) != 1 {
		t.Fatalf("Expected 1 error, got %#v", parser.errors)
	}
}

func TestCheckAwaitExpression(t *testing.T) {
	parser := MakeParser(nil)
	parser.scope.Add("req", Loc{}, makePromise(Number{}))
	expr := &UnaryExpression{
		Operator: token{kind: AwaitKeyword},
		Operand:  &Identifier{Token: literal{kind: Name, value: "req"}},
	}
	expr.typeCheck(parser)
	if len(parser.errors) != 0 {
		t.Fatalf("Expected no errors, got %#v", parser.errors)
	}
	if _, ok := expr.Type().(Number); !ok {
		t.Fatalf("Expected number type, got %v", expr.Type().Text())
	}
}

func TestCheckAwaitExpressionNotPromise(t *testing.T) {
	parser := MakeParser(nil)
	parser.scope.Add("req", Loc{}, Number{})
	expr := &UnaryExpression{
		Operator: token{kind: AwaitKeyword},
		Operand:  &Identifier{Token: literal{kind: Name, value: "req"}},
	}
	expr.typeCheck(parser)
	if len(parser.errors) != 1 {
		t.Fatalf("Expected 1 error, got %#v", parser.errors)
	}
	if _, ok := expr.Type().(Unknown); !ok {
		t.Fatalf("Expected unknown type, got %v", expr.Type().Text())
	}
}

func TestUnaryExpression(t *testing.T) {
	parser := MakeParser(strings.NewReader("?number"))
	expr := parser.parseUnaryExpression()

	if len(parser.errors) != 0 {
		t.Fatalf("Expected no errors, got %#v", parser.errors)
	}
	unary, ok := expr.(*UnaryExpression)
	if !ok {
		t.Fatal("Expected unary expression")
	}
	if unary.Operator.Kind() != QuestionMark {
		t.Fatal("Expected question mark")
	}
	if _, ok := unary.Operand.(*Literal); !ok {
		t.Fatal("Expected literal")
	}
	ty, ok := unary.Type().(Type)
	if !ok {
		t.Fatal("Expected type")
	}
	alias, ok := ty.Value.(TypeAlias)
	if !ok || alias.Name != "?" {
		t.Fatal("Expected option type")
	}
}

func TestCheckOptionType(t *testing.T) {
	// ?number
	parser := MakeParser(nil)
	expr := &UnaryExpression{
		Operator: token{kind: QuestionMark},
		Operand:  &Literal{Token: token{kind: NumberKeyword}},
	}
	expr.typeCheck(parser)

	if len(parser.errors) != 0 {
		t.Fatalf("Expected no errors, got %#v", parser.errors)
	}
	ty, ok := expr.Type().(Type)
	if !ok {
		t.Fatal("Expected type")
	}
	alias, ok := ty.Value.(TypeAlias)
	if !ok || alias.Name != "?" {
		t.Fatal("Expected option type")
	}
	if _, ok := getSomeType(alias.Ref.(Sum)).(Number); !ok {
		t.Fatal("Expected number option type")
	}
}

func TestNoOptionValue(t *testing.T) {
	parser := MakeParser(strings.NewReader("?42"))
	expr := parser.parseUnaryExpression()
	expr.typeCheck(parser)

	if len(parser.errors) != 1 {
		t.Fatalf("Expected 1 error, got %#v", parser.errors)
	}
}

func TestCheckErrorType(t *testing.T) {
	// !number
	parser := MakeParser(nil)
	expr := &UnaryExpression{
		Operator: token{kind: Bang},
		Operand:  &Literal{Token: token{kind: NumberKeyword}},
	}
	expr.typeCheck(parser)

	if len(parser.errors) != 0 {
		t.Fatalf("Expected no errors, got %#v", parser.errors)
	}
	ty, ok := expr.Type().(Type)
	if !ok {
		t.Fatal("Expected type")
	}
	alias, ok := ty.Value.(TypeAlias)
	if !ok || alias.Name != "!" {
		t.Fatal("Expected result type")
	}
	if _, ok := alias.Ref.(Sum).getMember("Ok").(Number); !ok {
		t.Fatal("Expected number option type")
	}
}

func TestCheckLogicalNot(t *testing.T) {
	// !true
	parser := MakeParser(nil)
	expr := &UnaryExpression{
		Operator: token{kind: Bang},
		Operand:  &Literal{literal{kind: BooleanLiteral, value: "true"}},
	}
	expr.typeCheck(parser)

	if len(parser.errors) != 0 {
		t.Fatalf("Expected no errors, got %#v", parser.errors)
	}
	if _, ok := expr.Type().(Boolean); !ok {
		t.Fatalf("Expected boolean")
	}

	// !42
	expr.Operand = &Literal{literal{kind: NumberLiteral, value: "42"}}
	expr.typeCheck(parser)
	if len(parser.errors) != 1 {
		t.Fatalf("Expected 1 error, got %#v", parser.errors)
	}
}

func TestParseReference(t *testing.T) {
	parser := MakeParser(strings.NewReader("&value"))
	node := parser.parseExpression()

	if len(parser.errors) != 0 {
		t.Fatalf("Expected no errors, got %+v: %#v", len(parser.errors), parser.errors)
	}

	if u, ok := node.(*UnaryExpression); !ok || u.Operator.Kind() != BinaryAnd {
		t.Fatalf("Expected '&' unary, got %#v", node)
	}
}

func TestParseReferenceBadOperand(t *testing.T) {
	parser := MakeParser(strings.NewReader("&value()"))
	parser.parseExpression()

	if len(parser.errors) != 1 {
		t.Fatalf("Expected 1 errors, got %+v: %#v", len(parser.errors), parser.errors)
	}
}

func TestCheckReference(t *testing.T) {
	parser := MakeParser(nil)
	// &number
	expr := &UnaryExpression{
		Operator: token{kind: BinaryAnd},
		Operand:  &Literal{token{kind: NumberKeyword}},
	}
	expr.typeCheck(parser)

	if len(parser.errors) != 0 {
		t.Fatalf("Expected no errors, got %#v", parser.errors)
	}
	if _, ok := expr.Type().(Type); !ok {
		t.Fatalf("Expected type, got %v", expr.Type().Text())
	}

	// &value
	parser.scope.Add("value", Loc{}, Number{})
	expr = &UnaryExpression{
		Operator: token{kind: BinaryAnd},
		Operand:  &Identifier{Token: literal{kind: Name, value: "value"}},
	}
	expr.typeCheck(parser)

	if len(parser.errors) != 0 {
		t.Fatalf("Expected no errors, got %#v", parser.errors)
	}
	if _, ok := expr.Type().(Ref); !ok {
		t.Fatalf("Expected ref, got %v", expr.Type().Text())
	}
}

func TestParseDereference(t *testing.T) {
	parser := MakeParser(strings.NewReader("*ref"))
	node := parser.parseExpression()

	if len(parser.errors) != 0 {
		t.Fatalf("Expected no errors, got %+v: %#v", len(parser.errors), parser.errors)
	}

	if u, ok := node.(*UnaryExpression); !ok || u.Operator.Kind() != Mul {
		t.Fatalf("Expected '*' unary, got %#v", node)
	}
}

func TestCheckDereference(t *testing.T) {
	parser := MakeParser(nil)
	// *ref
	parser.scope.Add("ref", Loc{}, Ref{Number{}})
	expr := &UnaryExpression{
		Operator: token{kind: Mul},
		Operand:  &Identifier{Token: literal{kind: Name, value: "ref"}},
	}
	expr.typeCheck(parser)

	if len(parser.errors) != 0 {
		t.Fatalf("Expected no errors, got %#v", parser.errors)
	}
	if _, ok := expr.Type().(Number); !ok {
		t.Fatalf("Expected number, got %v", expr.Type().Text())
	}

	expr.Operand = &Literal{literal{kind: NumberLiteral, value: "42"}}
	expr.typeCheck(parser)
	if len(parser.errors) != 1 {
		t.Fatalf("Expected 1 error, got %v: %#v", len(parser.errors), parser.errors)
	}
	if _, ok := expr.Type().(Unknown); !ok {
		t.Fatalf("Expected unknown, got %v", expr.Type().Text())
	}
}

func TestListTypeExpression(t *testing.T) {
	parser := MakeParser(strings.NewReader("[]number"))
	node := parser.parseExpression()

	if len(parser.errors) != 0 {
		t.Fatalf("Expected no errors, got %+v: %#v", len(parser.errors), parser.errors)
	}

	list, ok := node.(*ListTypeExpression)
	if !ok {
		t.Fatalf("Expected ListExpression, got %#v", node)
	}
	if list.Expr == nil {
		t.Fatalf("Expected a Type")
	}
}

func TestCheckListType(t *testing.T) {
	// []number
	parser := MakeParser(nil)
	expr := &ListTypeExpression{
		Expr: &Literal{Token: token{kind: NumberKeyword}},
	}
	expr.typeCheck(parser)

	if len(parser.errors) != 0 {
		t.Fatalf("Expected no errors, got %#v", parser.errors)
	}
	ty, ok := expr.Type().(Type)
	if !ok {
		t.Fatal("Expected type")
	}
	list, ok := ty.Value.(List)
	if !ok {
		t.Fatal("Expected list type")
	}
	if _, ok := list.Element.(Number); !ok {
		t.Fatal("Expected number list type")
	}
}

func TestCheckListTypeNoValue(t *testing.T) {
	// []42
	parser := MakeParser(nil)
	expr := &ListTypeExpression{
		Bracketed: &BracketedExpression{},
		Expr:      &Literal{Token: literal{kind: NumberLiteral, value: "42"}},
	}
	expr.typeCheck(parser)

	if len(parser.errors) != 1 {
		t.Fatalf("Expected 1 error, got %#v", parser.errors)
	}
	ty, ok := expr.Type().(Type)
	if !ok {
		t.Fatal("Expected type")
	}
	list, ok := ty.Value.(List)
	if !ok {
		t.Fatal("Expected list type")
	}
	if _, ok := list.Element.(Unknown); !ok {
		t.Fatal("Expected unknown list type")
	}
}

func TestNestedListTypeExpression(t *testing.T) {
	parser := MakeParser(strings.NewReader("[][]number"))
	node := parser.parseExpression()

	if len(parser.errors) != 0 {
		t.Fatalf("Expected no errors, got %+v: %#v", len(parser.errors), parser.errors)
	}

	list, ok := node.(*ListTypeExpression)
	if !ok {
		t.Fatalf("Expected ListExpression, got %#v", node)
	}
	if _, ok := list.Expr.(*ListTypeExpression); !ok {
		t.Fatalf("Expected a nested ListTypeExpression, got %#v", list.Type())
	}
	if list.Expr == nil {
		t.Fatalf("Expected a Type")
	}
}

func TestListTypeExpressionWithBracketed(t *testing.T) {
	parser := MakeParser(strings.NewReader("[number]number"))
	parser.parseExpression()

	if len(parser.errors) != 1 {
		t.Fatalf("Expected 1 error, got %#v", parser.errors)
	}
}

func TestParseTryExpression(t *testing.T) {
	parser := MakeParser(strings.NewReader("try result"))
	expr := parser.parseExpression()

	if len(parser.errors) > 0 {
		t.Fatalf("Expected no errors, got %#v", parser.errors)
	}
	_ = expr
}

func TestCheckTryExpression(t *testing.T) {
	parser := MakeParser(nil)
	parser.scope.Add("result", Loc{}, makeResultType(Number{}, nil))
	expr := &UnaryExpression{
		Operator: token{kind: TryKeyword},
		Operand:  &Identifier{Token: literal{kind: Name, value: "result"}},
	}
	expr.typeCheck(parser)

	if len(parser.errors) > 0 {
		t.Fatalf("Expected no errors, got %#v", parser.errors)
	}
	if _, ok := expr.Type().(Number); !ok {
		t.Fatalf("Expected a number, got %v", expr)
	}
}

func TestCheckTryExpressionBadType(t *testing.T) {
	parser := MakeParser(nil)
	expr := &UnaryExpression{
		Operator: token{kind: TryKeyword},
		Operand:  &Literal{literal{kind: NumberLiteral, value: "42"}},
	}
	expr.typeCheck(parser)

	if len(parser.errors) != 1 {
		t.Fatalf("Expected 1 error, got %#v", parser.errors)
	}
}
