package parser

import (
	"strings"
	"testing"
)

func TestParenthesized(t *testing.T) {
	parser := MakeParser(strings.NewReader("(42)"))
	paren := parser.parseParenthesizedExpression()
	if _, ok := paren.Expr.(*Literal); !ok {
		t.Fatalf("Expected literal between parentheses, got %v", paren.Expr)
	}
}

func TestParenthesizedTuple(t *testing.T) {
	parser := MakeParser(strings.NewReader("(1, 2)"))
	paren := parser.parseParenthesizedExpression()
	if _, ok := paren.Expr.(*TupleExpression); !ok {
		t.Fatalf("Expected TupleExpression between parentheses, got %#v", paren.Expr)
	}
}

func TestObjectDescriptionNoColon(t *testing.T) {
	parser := MakeParser(strings.NewReader("(n: number)"))
	parser.parseParenthesizedExpression()

	if len(parser.errors) != 0 {
		t.Fatalf("Expected no errors, got %#v", parser.errors)
	}
}
