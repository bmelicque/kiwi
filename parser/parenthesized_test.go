package parser

import "testing"

func TestParenthesized(t *testing.T) {
	tokenizer := testTokenizer{tokens: []Token{
		token{kind: LPAREN},
		literal{kind: NUMBER, value: "42"},
		token{kind: RPAREN},
	}}
	parser := MakeParser(&tokenizer)
	paren := parser.parseParenthesizedExpression()
	if _, ok := paren.Expr.(TokenExpression); !ok {
		t.Fatalf("Expected TokenExpression between parentheses, got %v", paren.Expr)
	}
}

func TestParenthesizedTuple(t *testing.T) {
	tokenizer := testTokenizer{tokens: []Token{
		token{kind: LPAREN},
		literal{kind: NUMBER, value: "1"},
		token{kind: COMMA},
		literal{kind: NUMBER, value: "2"},
		token{kind: RPAREN},
	}}
	parser := MakeParser(&tokenizer)
	paren := parser.parseParenthesizedExpression()
	if _, ok := paren.Expr.(TupleExpression); !ok {
		t.Fatalf("Expected TupleExpression between parentheses, got %#v", paren.Expr)
	}
}

func TestObjectDescriptionSingleLine(t *testing.T) {
	tokenizer := testTokenizer{tokens: []Token{
		token{kind: LPAREN},
		literal{kind: IDENTIFIER, value: "n"},
		token{kind: NUM_KW},
		token{kind: RPAREN},
	}}
	parser := MakeParser(&tokenizer)
	node := parser.parseParenthesizedExpression()

	if len(parser.errors) != 0 {
		t.Fatalf("Expected no errors, got %#v", parser.errors)
	}

	if _, ok := node.Expr.(TypedExpression); !ok {
		t.Fatalf("Expected TypedExpression, got %#v", node.Expr)
	}
}

func TestObjectDescription(t *testing.T) {
	tokenizer := testTokenizer{tokens: []Token{
		token{kind: LPAREN},
		token{kind: EOL},

		literal{kind: IDENTIFIER, value: "n"},
		token{kind: NUM_KW},
		token{kind: COMMA},
		token{kind: EOL},

		literal{kind: IDENTIFIER, value: "s"},
		token{kind: STR_KW},
		token{kind: COMMA},
		token{kind: EOL},

		token{kind: RPAREN},
	}}
	parser := MakeParser(&tokenizer)
	node := parser.parseParenthesizedExpression()

	if len(parser.errors) != 0 {
		t.Fatalf("Expected no errors, got %#v", parser.errors)
	}

	tuple, ok := node.Expr.(TupleExpression)
	if !ok {
		t.Fatalf("Expected TupleExpression, got %#v", node.Expr)
	}
	if len(tuple.Elements) != 2 {
		t.Fatalf("Expected 2 members, got %v", len(tuple.Elements))
	}
}

func TestObjectDescriptionNoColon(t *testing.T) {
	tokenizer := testTokenizer{tokens: []Token{
		token{kind: LPAREN},
		literal{kind: IDENTIFIER, value: "n"},
		token{kind: COLON},
		token{kind: NUM_KW},
		token{kind: RPAREN},
	}}
	parser := MakeParser(&tokenizer)
	parser.parseParenthesizedExpression()

	if len(parser.errors) != 0 {
		t.Fatalf("Expected no errors, got %#v", parser.errors)
	}
}
