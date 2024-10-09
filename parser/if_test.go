package parser

import "testing"

func TestIf(t *testing.T) {
	// if n == 2 { return 1 }
	tokenizer := testTokenizer{tokens: []Token{
		token{kind: IfKeyword},
		literal{kind: Name, value: "n"},
		token{kind: Equal},
		literal{kind: NumberLiteral, value: "2"},
		token{kind: LeftBrace},
		token{kind: ReturnKeyword},
		literal{kind: NumberLiteral, value: "1"},
		token{kind: RightBrace},
	}}
	parser := MakeParser(&tokenizer)
	node := parser.parseIfExpression()
	if node.Body == nil {
		t.Fatalf("Expected a body, got %#v", node)
	}
}

func TestIfElse(t *testing.T) {
	// if false {} else { true }
	tokenizer := testTokenizer{tokens: []Token{
		token{kind: IfKeyword},
		literal{kind: BooleanLiteral, value: "false"},
		token{kind: LeftBrace},
		token{kind: RightBrace},
		token{kind: ElseKeyword},
		token{kind: LeftBrace},
		literal{kind: BooleanLiteral, value: "true"},
		token{kind: RightBrace},
	}}
	parser := MakeParser(&tokenizer)
	node := parser.parseIfExpression()

	if len(parser.errors) != 0 {
		t.Fatalf("Expected no errors, got %#v", parser.errors)
	}

	if node.Body == nil {
		t.Fatal("Expected a body")
	}
	if node.Alternate == nil {
		t.Fatal("Expected alternate")
	}
	if _, ok := node.Alternate.(Block); !ok {
		t.Fatalf("Expected body alternate, got %#v", node.Alternate)
	}
}

func TestIfElseIf(t *testing.T) {
	// if false {} else if true {}
	tokenizer := testTokenizer{tokens: []Token{
		token{kind: IfKeyword},
		literal{kind: BooleanLiteral, value: "false"},
		token{kind: LeftBrace},
		token{kind: RightBrace},
		token{kind: ElseKeyword},
		token{kind: IfKeyword},
		literal{kind: BooleanLiteral, value: "true"},
		token{kind: LeftBrace},
		token{kind: RightBrace},
	}}
	parser := MakeParser(&tokenizer)
	node := parser.parseIfExpression()

	if len(parser.errors) != 0 {
		t.Fatalf("Expected no errors, got %#v", parser.errors)
	}

	if node.Body == nil {
		t.Fatal("Expected a body")
	}
	if node.Alternate == nil {
		t.Fatal("Expected alternate")
	}
	if _, ok := node.Alternate.(*IfExpression); !ok {
		t.Fatalf("Expected another 'if' as alternate, got %#v", node.Alternate)
	}
}
