package parser

import "testing"

func TestTuple(t *testing.T) {
	tokenizer := testTokenizer{tokens: []Token{
		literal{kind: NumberLiteral, value: "1"},
		token{kind: Comma},
		literal{kind: NumberLiteral, value: "2"},
		token{kind: Comma},
		literal{kind: NumberLiteral, value: "3"},
	}}
	parser := MakeParser(&tokenizer)
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
	tokenizer := testTokenizer{tokens: []Token{
		literal{kind: NumberLiteral, value: "1"},
		token{kind: NumberKeyword},
		token{kind: Comma},
		literal{kind: NumberLiteral, value: "2"},
		token{kind: NumberKeyword},
		token{kind: Comma},
		literal{kind: NumberLiteral, value: "3"},
		token{kind: NumberKeyword},
	}}
	parser := MakeParser(&tokenizer)
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

	if expr.Type().Kind() != NIL {
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

	if expr.Type().Kind() != STRING {
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
	if tuple.elements[0].Kind() != NUMBER {
		t.Fatalf("Expected number, got %#v", tuple.elements[0])
	}
	if tuple.elements[1].Kind() != STRING {
		t.Fatalf("Expected string, got %#v", tuple.elements[1])
	}
}
