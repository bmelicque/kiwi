package parser

import (
	"strings"
	"testing"
)

func TestParseAsync(t *testing.T) {
	parser, err := MakeParser(strings.NewReader("async fetch()"))
	if err != nil {
		t.Fatal(err)
	}
	parser.parseAsyncExpression()

	if len(parser.errors) > 0 {
		t.Fatalf("Expected no errors, got %#v", parser.errors)
	}
}

func TestParseAsyncNoExpr(t *testing.T) {
	parser, err := MakeParser(strings.NewReader("async"))
	if err != nil {
		t.Fatal(err)
	}
	parser.parseAsyncExpression()

	if len(parser.errors) != 1 {
		t.Fatalf("Expected 1 error, got %#v", parser.errors)
	}
}

func TestParseAsyncNotCall(t *testing.T) {
	parser, err := MakeParser(strings.NewReader("async fetch"))
	if err != nil {
		t.Fatal(err)
	}
	parser.parseAsyncExpression()

	if len(parser.errors) != 1 {
		t.Fatalf("Expected 1 error, got %#v", parser.errors)
	}
}

func TestCheckAsyncExpression(t *testing.T) {
	parser, _ := MakeParser(nil)
	parser.scope.Add("fetch", Loc{}, Function{
		Params:   &Tuple{},
		Returned: String{},
		Async:    true,
	})
	expr := &AsyncExpression{
		Keyword: token{kind: AsyncKeyword},
		Call: &CallExpression{
			Callee: &Identifier{Token: literal{kind: Name, value: "fetch"}},
			Args:   &ParenthesizedExpression{Expr: &TupleExpression{}},
		},
	}
	expr.typeCheck(parser)
	if len(parser.errors) != 0 {
		t.Fatalf("Expected no errors, got %#v", parser.errors)
	}
}
