package parser

import (
	"strings"
	"testing"
)

func TestParenthesized(t *testing.T) {
	parser, err := MakeParser(strings.NewReader("(42)"))
	if err != nil {
		t.Fatal(err)
	}
	paren := parser.parseParenthesizedExpression()
	if _, ok := paren.Expr.(*Literal); !ok {
		t.Fatalf("Expected literal between parentheses, got %v", paren.Expr)
	}
}

func TestParenthesizedTuple(t *testing.T) {
	parser, err := MakeParser(strings.NewReader("(1, 2)"))
	if err != nil {
		t.Fatal(err)
	}
	paren := parser.parseParenthesizedExpression()
	if _, ok := paren.Expr.(*TupleExpression); !ok {
		t.Fatalf("Expected TupleExpression between parentheses, got %#v", paren.Expr)
	}
}

func TestObjectDescriptionSingleLine(t *testing.T) {
	parser, err := MakeParser(strings.NewReader("(n number)"))
	if err != nil {
		t.Fatal(err)
	}
	node := parser.parseParenthesizedExpression()

	if len(parser.errors) != 0 {
		t.Fatalf("Expected no errors, got %#v", parser.errors)
	}

	if _, ok := node.Expr.(*Param); !ok {
		t.Fatalf("Expected TypedExpression, got %#v", node.Expr)
	}
}

func TestCheckObjectDescriptionSingleLine(t *testing.T) {
	parser, err := MakeParser(strings.NewReader(""))
	if err != nil {
		t.Fatal(err)
	}
	expr := &ParenthesizedExpression{Expr: &Param{
		Identifier: &Identifier{Token: literal{kind: Name, value: "value"}},
		Complement: &Literal{token{kind: NumberKeyword}},
	}}
	expr.typeCheck(parser)

	if len(parser.errors) != 0 {
		t.Fatalf("Expected no errors, got %#v", parser.errors)
	}

	typing, ok := expr.Type().(Type)
	if !ok {
		t.Fatalf("Expected type 'Type', got %#v", expr.Type())
	}
	object, ok := typing.Value.(Object)
	if !ok {
		t.Fatal("Expected an object")
	}
	_ = object
}

func TestObjectDescription(t *testing.T) {
	str := "(\n"
	str += "    n number,\n"
	str += "    s string,\n"
	str += ")"
	parser, err := MakeParser(strings.NewReader(str))
	if err != nil {
		t.Fatal(err)
	}
	node := parser.parseParenthesizedExpression()

	if len(parser.errors) != 0 {
		t.Fatalf("Expected no errors, got %#v", parser.errors)
	}

	tuple, ok := node.Expr.(*TupleExpression)
	if !ok {
		t.Fatalf("Expected TupleExpression, got %#v", node.Expr)
	}
	if len(tuple.Elements) != 2 {
		t.Fatalf("Expected 2 members, got %v", len(tuple.Elements))
	}
}

func TestObjectDescriptionNoColon(t *testing.T) {
	parser, err := MakeParser(strings.NewReader("(n: number)"))
	if err != nil {
		t.Fatal(err)
	}
	parser.parseParenthesizedExpression()

	if len(parser.errors) != 0 {
		t.Fatalf("Expected no errors, got %#v", parser.errors)
	}
}
