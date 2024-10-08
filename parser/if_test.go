package parser

import "testing"

func TestIf(t *testing.T) {
	// if n == 2 { return 1 }
	tokenizer := testTokenizer{tokens: []Token{
		token{kind: IF_KW},
		literal{kind: IDENTIFIER, value: "n"},
		token{kind: EQ},
		literal{kind: NUMBER, value: "2"},
		token{kind: LBRACE},
		token{kind: RETURN_KW},
		literal{kind: NUMBER, value: "1"},
		token{kind: RBRACE},
	}}
	parser := MakeParser(&tokenizer)
	node := parser.parseIf()
	statement, ok := node.(IfElse)
	if !ok {
		t.Fatalf("Expected 'if' statement, got %#v", node)
		return
	}
	if statement.Body == nil {
		t.Fatalf("Expected 'body' statement, got %#v", node)
	}
}

func TestIfElse(t *testing.T) {
	// if false {} else { true }
	tokenizer := testTokenizer{tokens: []Token{
		token{kind: IF_KW},
		literal{kind: BOOLEAN, value: "false"},
		token{kind: LBRACE},
		token{kind: RBRACE},
		token{kind: ELSE_KW},
		token{kind: LBRACE},
		literal{kind: BOOLEAN, value: "true"},
		token{kind: RBRACE},
	}}
	parser := MakeParser(&tokenizer)
	node := parser.parseIf()

	if len(parser.errors) != 0 {
		t.Fatalf("Expected no errors, got %#v", parser.errors)
	}

	statement, ok := node.(IfElse)
	if !ok {
		t.Fatalf("Expected 'if' statement, got %#v", node)
	}
	if statement.Body == nil {
		t.Fatal("Expected 'body' statement")
	}
	if statement.Alternate == nil {
		t.Fatal("Expected alternate")
	}
	if _, ok := statement.Alternate.(Block); !ok {
		t.Fatalf("Expected body alternate, got %#v", statement.Alternate)
	}
}

func TestIfElseIf(t *testing.T) {
	// if false {} else if true {}
	tokenizer := testTokenizer{tokens: []Token{
		token{kind: IF_KW},
		literal{kind: BOOLEAN, value: "false"},
		token{kind: LBRACE},
		token{kind: RBRACE},
		token{kind: ELSE_KW},
		token{kind: IF_KW},
		literal{kind: BOOLEAN, value: "true"},
		token{kind: LBRACE},
		token{kind: RBRACE},
	}}
	parser := MakeParser(&tokenizer)
	node := parser.parseIf()

	if len(parser.errors) != 0 {
		t.Fatalf("Expected no errors, got %#v", parser.errors)
	}

	statement, ok := node.(IfElse)
	if !ok {
		t.Fatalf("Expected 'if' statement, got %#v", node)
	}
	if statement.Body == nil {
		t.Fatal("Expected 'body' statement")
	}
	if statement.Alternate == nil {
		t.Fatal("Expected alternate")
	}
	if _, ok := statement.Alternate.(IfElse); !ok {
		t.Fatalf("Expected another 'if' as alternate, got %#v", statement.Alternate)
	}
}
