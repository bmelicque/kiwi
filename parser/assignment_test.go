package parser

import "testing"

func TestAssignment(t *testing.T) {
	tokenizer := testTokenizer{tokens: []Token{
		literal{kind: Name, value: "n"},
		token{kind: Assign},
		literal{kind: NumberLiteral, value: "42"},
	}}
	parser := MakeParser(&tokenizer)
	node := parser.parseAssignment()

	if len(parser.errors) != 0 {
		t.Fatalf("Expected no errors, got %#v", parser.errors)
	}

	expr, ok := node.(Assignment)
	if !ok {
		t.Fatalf("Expected Assignment, got %#v", node)
	}
	if _, ok := expr.Declared.(*Identifier); !ok {
		t.Fatalf("Expected token 'n'")
	}
	if _, ok := expr.Initializer.(*Literal); !ok {
		t.Fatalf("Expected literal 42")
	}
}

func TestTupleAssignment(t *testing.T) {
	tokenizer := testTokenizer{tokens: []Token{
		literal{kind: Name, value: "n"},
		token{kind: Comma},
		literal{kind: Name, value: "m"},
		token{kind: Assign},
		literal{kind: NumberLiteral, value: "1"},
		token{kind: Comma},
		literal{kind: NumberLiteral, value: "2"},
	}}
	parser := MakeParser(&tokenizer)
	node := parser.parseAssignment()

	expr, ok := node.(Assignment)
	if !ok {
		t.Fatalf("Expected Assignment, got %#v", node)
	}
	if _, ok := expr.Declared.(TupleExpression); !ok {
		t.Fatalf("Expected tuple 'n, m'")
	}
	if _, ok := expr.Initializer.(TupleExpression); !ok {
		t.Fatalf("Expected tuple 'n, m'")
	}
}

func TestObjectDeclaration(t *testing.T) {
	tokenizer := testTokenizer{tokens: []Token{
		literal{kind: Name, value: "Type"},
		token{kind: Define},
		token{kind: LeftParenthesis},
		token{kind: EOL},

		literal{kind: Name, value: "n"},
		token{kind: NumberKeyword},
		token{kind: Comma},
		token{kind: EOL},

		token{kind: RightParenthesis},
	}}
	parser := MakeParser(&tokenizer)
	node := parser.parseAssignment()

	expr, ok := node.(Assignment)
	if !ok {
		t.Fatalf("Expected Assignment, got %#v", node)
	}
	if _, ok := expr.Declared.(*Identifier); !ok {
		t.Fatalf("Expected identifier 'Type'")
	}
	if _, ok := expr.Initializer.(ParenthesizedExpression); !ok {
		t.Fatalf("Expected ParenthesizedExpression")
	}
}

func TestMethodDeclaration(t *testing.T) {
	tokenizer := testTokenizer{tokens: []Token{
		token{kind: LeftParenthesis},
		literal{kind: Name, value: "t"},
		literal{kind: Name, value: "Type"},
		token{kind: RightParenthesis},
		token{kind: Dot},
		literal{kind: Name, value: "method"},
		token{kind: Define},
		token{kind: LeftParenthesis},
		token{kind: RightParenthesis},
		token{kind: SlimArrow},
		token{kind: LeftParenthesis},
		token{kind: RightParenthesis},
	}}
	parser := MakeParser(&tokenizer)
	node := parser.parseAssignment()

	expr, ok := node.(Assignment)
	if !ok {
		t.Fatalf("Expected Assignment, got %#v", node)
	}
	if _, ok := expr.Declared.(PropertyAccessExpression); !ok {
		t.Fatalf("Expected method declaration")
	}
	if _, ok := expr.Initializer.(FunctionExpression); !ok {
		t.Fatalf("Expected FunctionExpression")
	}
}
