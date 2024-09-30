package parser

import (
	"testing"

	"github.com/bmelicque/test-parser/tokenizer"
)

func TestParenthesized(t *testing.T) {
	tokenizer := testTokenizer{tokens: []tokenizer.Token{
		testToken{kind: tokenizer.LPAREN},
		testToken{kind: tokenizer.NUMBER, value: "42"},
		testToken{kind: tokenizer.RPAREN},
	}}
	parser := MakeParser(&tokenizer)
	paren := parser.parseParenthesizedExpression()
	if _, ok := paren.Expr.(TokenExpression); !ok {
		t.Fatalf("Expected TokenExpression between parentheses, got %v", paren.Expr)
	}
}

func TestParenthesizedTuple(t *testing.T) {
	tokenizer := testTokenizer{tokens: []tokenizer.Token{
		testToken{kind: tokenizer.LPAREN},
		testToken{kind: tokenizer.NUMBER, value: "1"},
		testToken{kind: tokenizer.COMMA},
		testToken{kind: tokenizer.NUMBER, value: "2"},
		testToken{kind: tokenizer.RPAREN},
	}}
	parser := MakeParser(&tokenizer)
	paren := parser.parseParenthesizedExpression()
	if _, ok := paren.Expr.(TupleExpression); !ok {
		t.Fatalf("Expected TupleExpression between parentheses, got %#v", paren.Expr)
	}
}

func TestObjectDescriptionSingleLine(t *testing.T) {
	tokenizer := testTokenizer{tokens: []tokenizer.Token{
		testToken{kind: tokenizer.LPAREN},
		testToken{kind: tokenizer.IDENTIFIER, value: "n"},
		testToken{kind: tokenizer.NUM_KW},
		testToken{kind: tokenizer.RPAREN},
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
	tokenizer := testTokenizer{tokens: []tokenizer.Token{
		testToken{kind: tokenizer.LPAREN},
		testToken{kind: tokenizer.EOL},

		testToken{kind: tokenizer.IDENTIFIER, value: "n"},
		testToken{kind: tokenizer.NUM_KW},
		testToken{kind: tokenizer.COMMA},
		testToken{kind: tokenizer.EOL},

		testToken{kind: tokenizer.IDENTIFIER, value: "s"},
		testToken{kind: tokenizer.STR_KW},
		testToken{kind: tokenizer.COMMA},
		testToken{kind: tokenizer.EOL},

		testToken{kind: tokenizer.RPAREN},
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
	tokenizer := testTokenizer{tokens: []tokenizer.Token{
		testToken{kind: tokenizer.LPAREN},
		testToken{kind: tokenizer.IDENTIFIER, value: "n"},
		testToken{kind: tokenizer.COLON},
		testToken{kind: tokenizer.NUM_KW},
		testToken{kind: tokenizer.RPAREN},
	}}
	parser := MakeParser(&tokenizer)
	parser.parseParenthesizedExpression()

	if len(parser.errors) != 0 {
		t.Fatalf("Expected no errors, got %#v", parser.errors)
	}
}
