package parser

import (
	"strings"
	"testing"
)

func TestTuple(t *testing.T) {
	parser := MakeParser(strings.NewReader("1, 2, 3"))
	node := parser.parseTupleExpression()

	if len(parser.errors) > 0 {
		t.Fatalf("Expected no errors, got %#v", parser.errors)
	}

	tuple, ok := node.(*TupleExpression)
	if !ok {
		t.Fatalf("Expected TupleExpression, got %#v", node)
		return
	}
	if len(tuple.Elements) != 3 {
		t.Fatalf("Expected 3 elements, got %v", len(tuple.Elements))
	}
}

func TestTypedTuple(t *testing.T) {
	str := "a number, b number, c number"
	parser := MakeParser(strings.NewReader(str))
	node := parser.parseTupleExpression()

	if len(parser.errors) > 0 {
		t.Fatalf("Expected no errors, got %#v", parser.errors)
	}

	tuple, ok := node.(*TupleExpression)
	if !ok {
		t.Fatalf("Expected TupleExpression, got %#v", node)
		return
	}
	if len(tuple.Elements) != 3 {
		t.Fatalf("Expected 3 elements, got %v", len(tuple.Elements))
	}
}

func TestEmptyTupleType(t *testing.T) {
	parser := MakeParser(nil)
	expr := &TupleExpression{}
	expr.typeCheck(parser)

	if len(parser.errors) > 0 {
		t.Fatalf("Expected no errors, got %#v", parser.errors)
	}

	if _, ok := expr.Type().(Nil); !ok {
		t.Fatalf("Expected nil type, got %#v", expr.Type())
	}
}

func TestSingleTupleType(t *testing.T) {
	parser := MakeParser(nil)
	expr := &TupleExpression{Elements: []Expression{
		&Literal{Token: literal{kind: StringLiteral, value: "\"Hi!\""}},
	}}
	expr.typeCheck(parser)

	if len(parser.errors) > 0 {
		t.Fatalf("Expected no errors, got %#v", parser.errors)
	}

	if _, ok := expr.Type().(String); !ok {
		t.Fatalf("Expected string type, got %#v", expr.Type())
	}
}

func TestTupleType(t *testing.T) {
	parser := MakeParser(nil)
	expr := &TupleExpression{Elements: []Expression{
		&Literal{Token: literal{kind: NumberLiteral, value: "42"}},
		&Literal{Token: literal{kind: StringLiteral, value: "\"Hi!\""}},
	}}
	expr.typeCheck(parser)

	if len(parser.errors) > 0 {
		t.Fatalf("Expected no errors, got %#v", parser.errors)
	}

	tuple, ok := expr.Type().(Tuple)
	if !ok {
		t.Fatalf("Expected tuple type, got %#v", expr.Type())
	}
	if _, ok := tuple.Elements[0].(Number); !ok {
		t.Fatalf("Expected number, got %#v", tuple.Elements[0])
	}
	if _, ok := tuple.Elements[1].(String); !ok {
		t.Fatalf("Expected string, got %#v", tuple.Elements[1])
	}
}
