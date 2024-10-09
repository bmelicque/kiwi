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
	alias, ok := node.Type().(TypeAlias)
	if !ok || alias.Name != "Option" {
		t.Fatalf("Expected option type")
	}
}

func TestIfWithNonBoolean(t *testing.T) {
	// if 42 { }
	tokenizer := testTokenizer{tokens: []Token{
		token{kind: IfKeyword},
		literal{kind: NumberLiteral, value: "42"},
		token{kind: LeftBrace},
		token{kind: RightBrace},
	}}
	parser := MakeParser(&tokenizer)
	parser.parseIfExpression()
	if len(parser.errors) != 1 {
		t.Fatalf("Expected 1 error, got %#v", parser.errors)
	}
}

func TestIfElse(t *testing.T) {
	// if false { true } else { false }
	tokenizer := testTokenizer{tokens: []Token{
		token{kind: IfKeyword},
		literal{kind: BooleanLiteral, value: "false"},
		token{kind: LeftBrace},
		literal{kind: BooleanLiteral, value: "true"},
		token{kind: RightBrace},
		token{kind: ElseKeyword},
		token{kind: LeftBrace},
		literal{kind: BooleanLiteral, value: "false"},
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
	if node.Type().Kind() != BOOLEAN {
		t.Fatalf("Expected a boolean")
	}
}

func TestIfElseWithTypeMismatch(t *testing.T) {
	// if false { 42 } else { false }
	tokenizer := testTokenizer{tokens: []Token{
		token{kind: IfKeyword},
		literal{kind: BooleanLiteral, value: "false"},
		token{kind: LeftBrace},
		literal{kind: NumberLiteral, value: "42"},
		token{kind: RightBrace},
		token{kind: ElseKeyword},
		token{kind: LeftBrace},
		literal{kind: BooleanLiteral, value: "false"},
		token{kind: RightBrace},
	}}
	parser := MakeParser(&tokenizer)
	parser.parseIfExpression()

	if len(parser.errors) != 1 {
		t.Fatalf("Expected 1 error, got %#v", parser.errors)
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
