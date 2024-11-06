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
