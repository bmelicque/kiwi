package parser

import (
	"testing"

	"github.com/bmelicque/test-parser/tokenizer"
)

func TestIf(t *testing.T) {
	// if n == 2 { return 1 }
	tokenizer := testTokenizer{tokens: []tokenizer.Token{
		testToken{kind: tokenizer.IF_KW},
		testToken{kind: tokenizer.IDENTIFIER, value: "n"},
		testToken{kind: tokenizer.EQ},
		testToken{kind: tokenizer.NUMBER, value: "2"},
		testToken{kind: tokenizer.LBRACE},
		testToken{kind: tokenizer.RETURN_KW},
		testToken{kind: tokenizer.NUMBER, value: "1"},
		testToken{kind: tokenizer.RBRACE},
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
	tokenizer := testTokenizer{tokens: []tokenizer.Token{
		testToken{kind: tokenizer.IF_KW},
		testToken{kind: tokenizer.BOOLEAN, value: "false"},
		testToken{kind: tokenizer.LBRACE},
		testToken{kind: tokenizer.RBRACE},
		testToken{kind: tokenizer.ELSE_KW},
		testToken{kind: tokenizer.LBRACE},
		testToken{kind: tokenizer.BOOLEAN, value: "true"},
		testToken{kind: tokenizer.RBRACE},
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
	if _, ok := statement.Alternate.(Body); !ok {
		t.Fatalf("Expected body alternate, got %#v", statement.Alternate)
	}
}

func TestIfElseIf(t *testing.T) {
	// if false {} else if true {}
	tokenizer := testTokenizer{tokens: []tokenizer.Token{
		testToken{kind: tokenizer.IF_KW},
		testToken{kind: tokenizer.BOOLEAN, value: "false"},
		testToken{kind: tokenizer.LBRACE},
		testToken{kind: tokenizer.RBRACE},
		testToken{kind: tokenizer.ELSE_KW},
		testToken{kind: tokenizer.IF_KW},
		testToken{kind: tokenizer.BOOLEAN, value: "true"},
		testToken{kind: tokenizer.LBRACE},
		testToken{kind: tokenizer.RBRACE},
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
