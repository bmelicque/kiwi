package parser

import "testing"

func TestListTypeExpression(t *testing.T) {
	tokenizer := testTokenizer{tokens: []Token{
		token{kind: LBRACKET},
		token{kind: RBRACKET},
		token{kind: NUM_KW},
	}}
	parser := MakeParser(&tokenizer)
	node := ParseExpression(parser)

	if len(parser.errors) != 0 {
		t.Fatalf("Expected no errors, got %+v: %#v", len(parser.errors), parser.errors)
	}

	list, ok := node.(ListTypeExpression)
	if !ok {
		t.Fatalf("Expected ListExpression, got %#v", node)
	}
	if list.Type == nil {
		t.Fatalf("Expected a Type")
	}
}

func TestNestedListTypeExpression(t *testing.T) {
	tokenizer := testTokenizer{tokens: []Token{
		token{kind: LBRACKET},
		token{kind: RBRACKET},
		token{kind: LBRACKET},
		token{kind: RBRACKET},
		token{kind: NUM_KW},
	}}
	parser := MakeParser(&tokenizer)
	node := ParseExpression(parser)

	if len(parser.errors) != 0 {
		t.Fatalf("Expected no errors, got %+v: %#v", len(parser.errors), parser.errors)
	}

	list, ok := node.(ListTypeExpression)
	if !ok {
		t.Fatalf("Expected ListExpression, got %#v", node)
	}
	if _, ok := list.Type.(ListTypeExpression); !ok {
		t.Fatalf("Expected a nested ListTypeExpression, got %#v", list.Type)
	}
	if list.Type == nil {
		t.Fatalf("Expected a Type")
	}
}
