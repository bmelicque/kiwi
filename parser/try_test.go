package parser

import "testing"

func TestParseTryExpression(t *testing.T) {
	parser := MakeParser(&testTokenizer{tokens: []Token{
		token{kind: TryKeyword},
		literal{kind: Name, value: "result"},
	}})
	expr := parser.parseTryExpression()

	if len(parser.errors) > 0 {
		t.Fatalf("Expected no errors, got %#v", parser.errors)
	}
	_ = expr
}

func TestCheckTryExpression(t *testing.T) {
	parser := MakeParser(nil)
	parser.scope.Add("result", Loc{}, makeResultType(Number{}, nil))
	expr := &TryExpression{
		Expr: &Identifier{Token: literal{kind: Name, value: "result"}},
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
	expr := &TryExpression{
		Expr: &Literal{literal{kind: NumberLiteral, value: "42"}},
	}
	expr.typeCheck(parser)

	if len(parser.errors) != 1 {
		t.Fatalf("Expected 1 error, got %#v", parser.errors)
	}
}
