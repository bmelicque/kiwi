package parser

import "testing"

func TestSumType(t *testing.T) {
	tokenizer := testTokenizer{tokens: []Token{
		token{kind: BinaryOr},
		literal{kind: Name, value: "Some"},
		token{kind: LeftParenthesis},
		literal{kind: Name, value: "Type"},
		token{kind: RightParenthesis},
		token{kind: BinaryOr},
		literal{kind: Name, value: "None"},
	}}
	parser := MakeParser(&tokenizer)
	node := parser.parseSumType()

	if len(parser.errors) > 0 {
		t.Fatalf("Expected no errors, got %v: %#v", len(parser.errors), parser.errors)
	}

	sum, ok := node.(*SumType)
	if !ok {
		t.Fatalf("Expected SumType, got %#v", node)
		return
	}
	if len(sum.Members) != 2 {
		t.Fatalf("Expected 2 elements, got %v: %#v", len(sum.Members), sum.Members)
	}
}

func TestSumTypeLength(t *testing.T) {
	tokenizer := testTokenizer{tokens: []Token{
		token{kind: BinaryOr},
		literal{kind: Name, value: "Alone"},
	}}
	parser := MakeParser(&tokenizer)
	parser.parseSumType()

	if len(parser.errors) != 1 {
		t.Fatalf("Expected 1 error, got %#v", parser.errors)
	}
}
