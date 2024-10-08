package parser

import "testing"

func TestAssignment(t *testing.T) {
	tokenizer := testTokenizer{tokens: []Token{
		literal{kind: IDENTIFIER, value: "n"},
		token{kind: ASSIGN},
		literal{kind: NUMBER, value: "42"},
	}}
	parser := MakeParser(&tokenizer)
	node := parser.parseAssignment()

	expr, ok := node.(Assignment)
	if !ok {
		t.Fatalf("Expected Assignment, got %#v", node)
	}
	if _, ok := expr.Declared.(TokenExpression); !ok {
		t.Fatalf("Expected token 'n'")
	}
	if _, ok := expr.Initializer.(TokenExpression); !ok {
		t.Fatalf("Expected literal 42")
	}
}

func TestTupleAssignment(t *testing.T) {
	tokenizer := testTokenizer{tokens: []Token{
		literal{kind: IDENTIFIER, value: "n"},
		token{kind: COMMA},
		literal{kind: IDENTIFIER, value: "m"},
		token{kind: ASSIGN},
		literal{kind: NUMBER, value: "1"},
		token{kind: COMMA},
		literal{kind: NUMBER, value: "2"},
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
		literal{kind: IDENTIFIER, value: "Type"},
		token{kind: DEFINE},
		token{kind: LPAREN},
		token{kind: EOL},

		literal{kind: IDENTIFIER, value: "n"},
		token{kind: NUM_KW},
		token{kind: COMMA},
		token{kind: EOL},

		token{kind: RPAREN},
	}}
	parser := MakeParser(&tokenizer)
	node := parser.parseAssignment()

	expr, ok := node.(Assignment)
	if !ok {
		t.Fatalf("Expected Assignment, got %#v", node)
	}
	if _, ok := expr.Declared.(TokenExpression); !ok {
		t.Fatalf("Expected identifier 'Type'")
	}
	if _, ok := expr.Initializer.(ParenthesizedExpression); !ok {
		t.Fatalf("Expected ParenthesizedExpression")
	}
}

func TestMethodDeclaration(t *testing.T) {
	tokenizer := testTokenizer{tokens: []Token{
		token{kind: LPAREN},
		literal{kind: IDENTIFIER, value: "t"},
		literal{kind: IDENTIFIER, value: "Type"},
		token{kind: RPAREN},
		token{kind: DOT},
		literal{kind: IDENTIFIER, value: "method"},
		token{kind: DEFINE},
		token{kind: LPAREN},
		token{kind: RPAREN},
		token{kind: SLIM_ARR},
		token{kind: LPAREN},
		token{kind: RPAREN},
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
