package parser

import "testing"

func TestParenthesized(t *testing.T) {
	tokenizer := testTokenizer{tokens: []Token{
		token{kind: LeftParenthesis},
		literal{kind: NumberLiteral, value: "42"},
		token{kind: RightParenthesis},
	}}
	parser := MakeParser(&tokenizer)
	paren := parser.parseParenthesizedExpression()
	if _, ok := paren.Expr.(*Literal); !ok {
		t.Fatalf("Expected literal between parentheses, got %v", paren.Expr)
	}
}

func TestParenthesizedTuple(t *testing.T) {
	tokenizer := testTokenizer{tokens: []Token{
		token{kind: LeftParenthesis},
		literal{kind: NumberLiteral, value: "1"},
		token{kind: Comma},
		literal{kind: NumberLiteral, value: "2"},
		token{kind: RightParenthesis},
	}}
	parser := MakeParser(&tokenizer)
	paren := parser.parseParenthesizedExpression()
	if _, ok := paren.Expr.(*TupleExpression); !ok {
		t.Fatalf("Expected TupleExpression between parentheses, got %#v", paren.Expr)
	}
}

func TestObjectDescriptionSingleLine(t *testing.T) {
	tokenizer := testTokenizer{tokens: []Token{
		token{kind: LeftParenthesis},
		literal{kind: Name, value: "n"},
		token{kind: NumberKeyword},
		token{kind: RightParenthesis},
	}}
	parser := MakeParser(&tokenizer)
	node := parser.parseParenthesizedExpression()

	if len(parser.errors) != 0 {
		t.Fatalf("Expected no errors, got %#v", parser.errors)
	}

	if _, ok := node.Expr.(*TypedExpression); !ok {
		t.Fatalf("Expected TypedExpression, got %#v", node.Expr)
	}
}

func TestObjectDescription(t *testing.T) {
	tokenizer := testTokenizer{tokens: []Token{
		token{kind: LeftParenthesis},
		token{kind: EOL},

		literal{kind: Name, value: "n"},
		token{kind: NumberKeyword},
		token{kind: Comma},
		token{kind: EOL},

		literal{kind: Name, value: "s"},
		token{kind: StringKeyword},
		token{kind: Comma},
		token{kind: EOL},

		token{kind: RightParenthesis},
	}}
	parser := MakeParser(&tokenizer)
	node := parser.parseParenthesizedExpression()

	if len(parser.errors) != 0 {
		t.Fatalf("Expected no errors, got %#v", parser.errors)
	}

	tuple, ok := node.Expr.(*TupleExpression)
	if !ok {
		t.Fatalf("Expected TupleExpression, got %#v", node.Expr)
	}
	if len(tuple.Elements) != 2 {
		t.Fatalf("Expected 2 members, got %v", len(tuple.Elements))
	}
}

func TestObjectDescriptionNoColon(t *testing.T) {
	tokenizer := testTokenizer{tokens: []Token{
		token{kind: LeftParenthesis},
		literal{kind: Name, value: "n"},
		token{kind: Colon},
		token{kind: NumberKeyword},
		token{kind: RightParenthesis},
	}}
	parser := MakeParser(&tokenizer)
	parser.parseParenthesizedExpression()

	if len(parser.errors) != 0 {
		t.Fatalf("Expected no errors, got %#v", parser.errors)
	}
}
