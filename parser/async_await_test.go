package parser

import "testing"

func TestParseAsync(t *testing.T) {
	parser := MakeParser(&testTokenizer{tokens: []Token{
		token{kind: AsyncKeyword},
		literal{kind: Name, value: "fetch"},
		token{kind: LeftParenthesis},
		token{kind: RightParenthesis},
	}})
	parser.parseAsyncExpression()

	if len(parser.errors) > 0 {
		t.Fatalf("Expected no errors, got %#v", parser.errors)
	}
}

func TestParseAsyncNoExpr(t *testing.T) {
	parser := MakeParser(&testTokenizer{tokens: []Token{
		token{kind: AsyncKeyword},
	}})
	parser.parseAsyncExpression()

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
	expr := &AsyncExpression{
		Keyword: token{kind: AsyncKeyword},
		Expr: &CallExpression{
			Callee: &Identifier{Token: literal{kind: Name, value: "fetch"}},
			Args:   &ParenthesizedExpression{Expr: &TupleExpression{}},
		},
	}
	expr.typeCheck(parser)
	if len(parser.errors) != 0 {
		t.Fatalf("Expected no errors, got %#v", parser.errors)
	}
}

func TestCheckAsyncExpressionOnlyFunctionCalls(t *testing.T) {
	parser := MakeParser(nil)
	parser.scope.Add("fetch", Loc{}, Function{
		Params:   &Tuple{},
		Returned: String{},
		Async:    true,
	})
	expr := &AsyncExpression{
		Keyword: token{kind: AsyncKeyword},
		Expr:    &Identifier{Token: literal{kind: Name, value: "fetch"}},
	}
	expr.typeCheck(parser)
	if len(parser.errors) != 1 {
		t.Fatalf("Expected 1 error, got %#v", parser.errors)
	}
}

func TestParseAwaitExpression(t *testing.T) {
	parser := MakeParser(&testTokenizer{tokens: []Token{
		token{kind: AwaitKeyword},
		literal{kind: Name, value: "request"},
	}})
	parser.parseAwaitExpression()

	if len(parser.errors) > 0 {
		t.Fatalf("Expected no errors, got %#v", parser.errors)
	}
}

func TestParseAwaitExpressionNoExpr(t *testing.T) {
	parser := MakeParser(&testTokenizer{tokens: []Token{
		token{kind: AwaitKeyword},
	}})
	parser.parseAwaitExpression()

	if len(parser.errors) != 1 {
		t.Fatalf("Expected 1 error, got %#v", parser.errors)
	}
}

func TestCheckAwaitExpression(t *testing.T) {
	parser := MakeParser(nil)
	parser.scope.Add("req", Loc{}, makePromise(Number{}))
	expr := &AwaitExpression{
		Keyword: token{kind: AsyncKeyword},
		Expr:    &Identifier{Token: literal{kind: Name, value: "req"}},
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
	expr := &AwaitExpression{
		Keyword: token{kind: AsyncKeyword},
		Expr:    &Identifier{Token: literal{kind: Name, value: "req"}},
	}
	expr.typeCheck(parser)
	if len(parser.errors) != 1 {
		t.Fatalf("Expected 1 error, got %#v", parser.errors)
	}
	if _, ok := expr.Type().(Unknown); !ok {
		t.Fatalf("Expected unknown type, got %v", expr.Type().Text())
	}
}
