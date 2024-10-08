package parser

import "testing"

func TestUnaryExpression(t *testing.T) {
	parser := MakeParser(&testTokenizer{tokens: []Token{
		token{kind: QuestionMark},
		token{kind: NumberKeyword},
	}})
	expr := parser.parseUnaryExpression()

	if len(parser.errors) != 0 {
		t.Fatalf("Expected no errors, got %#v", parser.errors)
	}
	unary, ok := expr.(*UnaryExpression)
	if !ok {
		t.Fatal("Expected unary expression")
	}
	if unary.Operator.Kind() != QuestionMark {
		t.Fatal("Expected question mark")
	}
	if _, ok := unary.Operand.(Literal); !ok {
		t.Fatal("Expected literal")
	}
}

func TestNoOptionValue(t *testing.T) {
	parser := MakeParser(&testTokenizer{tokens: []Token{
		token{kind: QuestionMark},
		literal{kind: NumberLiteral, value: "42"},
	}})
	parser.parseUnaryExpression()

	if len(parser.errors) != 1 {
		t.Fatalf("Expected 1 error, got %#v", parser.errors)
	}
}

func TestListTypeExpression(t *testing.T) {
	tokenizer := testTokenizer{tokens: []Token{
		token{kind: LeftBracket},
		token{kind: RightBracket},
		token{kind: NumberKeyword},
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
		token{kind: LeftBracket},
		token{kind: RightBracket},
		token{kind: LeftBracket},
		token{kind: RightBracket},
		token{kind: NumberKeyword},
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
