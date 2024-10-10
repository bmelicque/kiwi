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

	tuple, ok := node.(*TupleExpression)
	if !ok {
		t.Fatalf("Expected TupleExpression, got %#v", node)
		return
	}
	if len(tuple.Elements) != 3 {
		t.Fatalf("Expected 3 elements, got %v", len(tuple.Elements))
	}
}
