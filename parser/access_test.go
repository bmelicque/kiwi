package parser

import "testing"

func TestComputedPropertyAccess(t *testing.T) {
	tokenizer := testTokenizer{tokens: []Token{
		literal{kind: IDENTIFIER, value: "n"},
		token{kind: LBRACKET},
		literal{kind: IDENTIFIER, value: "p"},
		token{kind: RBRACKET},
	}}
	parser := MakeParser(&tokenizer)
	node := parser.parseAccessExpression()

	expr, ok := node.(ComputedAccessExpression)
	if !ok {
		t.Fatalf("Expected ComputedAccessExpression, got %#v", node)
	}
	if _, ok := expr.Expr.(TokenExpression); !ok {
		t.Fatalf("Expected token 'n'")
	}
	if _, ok := expr.Property.Expr.(TokenExpression); !ok {
		t.Fatalf("Expected token 'p'")
	}
}

func TestPropertyAccess(t *testing.T) {
	tokenizer := testTokenizer{tokens: []Token{
		literal{kind: IDENTIFIER, value: "n"},
		token{kind: DOT},
		literal{kind: IDENTIFIER, value: "p"},
	}}
	parser := MakeParser(&tokenizer)
	node := parser.parseAccessExpression()

	expr, ok := node.(PropertyAccessExpression)
	if !ok {
		t.Fatalf("Expected PropertyAccessExpression, got %#v", node)
	}
	if _, ok := expr.Expr.(TokenExpression); !ok {
		t.Fatalf("Expected token 'n'")
	}
	if _, ok := expr.Property.(TokenExpression); !ok {
		t.Fatalf("Expected token 'p'")
	}
}

func TestTupleAccess(t *testing.T) {
	tokenizer := testTokenizer{tokens: []Token{
		literal{kind: IDENTIFIER, value: "tuple"},
		token{kind: DOT},
		literal{kind: NUMBER, value: "0"},
	}}
	parser := MakeParser(&tokenizer)
	node := parser.parseAccessExpression()

	expr, ok := node.(PropertyAccessExpression)
	if !ok {
		t.Fatalf("Expected PropertyAccessExpression, got %#v", node)
	}
	if _, ok := expr.Expr.(TokenExpression); !ok {
		t.Fatalf("Expected token 'n'")
	}
	if _, ok := expr.Property.(TokenExpression); !ok {
		t.Fatalf("Expected token 'p'")
	}
}

func TestMethodAccess(t *testing.T) {
	tokenizer := testTokenizer{tokens: []Token{
		token{kind: LPAREN},
		literal{kind: IDENTIFIER, value: "t"},
		literal{kind: IDENTIFIER, value: "Type"},
		token{kind: RPAREN},
		token{kind: DOT},
		literal{kind: IDENTIFIER, value: "method"},
	}}
	parser := MakeParser(&tokenizer)
	node := parser.parseAccessExpression()

	expr, ok := node.(PropertyAccessExpression)
	if !ok {
		t.Fatalf("Expected PropertyAccessExpression, got %#v", node)
	}

	if _, ok := expr.Expr.(ParenthesizedExpression); !ok {
		t.Fatalf("Expected ParenthesizedExpression on LHS, got %#v", expr.Expr)
	}

	if _, ok := expr.Property.(TokenExpression); !ok {
		t.Fatalf("Expected token 'method'")
	}
}

func TestTraitDefinition(t *testing.T) {
	tokenizer := testTokenizer{tokens: []Token{
		token{kind: LPAREN},
		literal{kind: IDENTIFIER, value: "Self"},
		token{kind: RPAREN},
		token{kind: DOT},
		token{kind: LPAREN},
		literal{kind: IDENTIFIER, value: "method"},
		token{kind: LPAREN},
		token{kind: RPAREN},
		token{kind: SLIM_ARR},
		literal{kind: IDENTIFIER, value: "Self"},
		token{kind: RPAREN},
	}}

	parser := MakeParser(&tokenizer)
	node := parser.parseAccessExpression()

	if len(parser.errors) > 0 {
		t.Fatalf("Got %v parsing errors: %#v", len(parser.errors), parser.errors)
	}

	expr, ok := node.(PropertyAccessExpression)
	if !ok {
		t.Fatalf("Expected PropertyAccessExpression, got %#v", node)
	}

	if _, ok := expr.Expr.(ParenthesizedExpression); !ok {
		t.Fatalf("Expected ParenthesizedExpression, got %#v", expr.Expr)
	}

	if _, ok := expr.Property.(ParenthesizedExpression); !ok {
		t.Fatalf("Expected ParenthesizedExpression, got %#v", expr.Property)
	}
}

func TestFunctionCall(t *testing.T) {
	tokenizer := testTokenizer{tokens: []Token{
		literal{kind: IDENTIFIER, value: "f"},
		token{kind: LPAREN},
		literal{kind: NUMBER, value: "42"},
		token{kind: RPAREN},
	}}
	parser := MakeParser(&tokenizer)
	node := parser.parseAccessExpression()

	expr, ok := node.(CallExpression)
	if !ok {
		t.Fatalf("Expected CallExpression, got %#v", node)
	}
	if _, ok := expr.Callee.(TokenExpression); !ok {
		t.Fatalf("Expected token 'f'")
	}
}

func TestFunctionCallWithTypeArgs(t *testing.T) {
	tokenizer := testTokenizer{tokens: []Token{
		literal{kind: IDENTIFIER, value: "f"},
		token{kind: LBRACKET},
		token{kind: NUM_KW},
		token{kind: RBRACKET},
		token{kind: LPAREN},
		literal{kind: NUMBER, value: "42"},
		token{kind: RPAREN},
	}}
	parser := MakeParser(&tokenizer)
	node := parser.parseAccessExpression()

	expr, ok := node.(CallExpression)
	if !ok {
		t.Fatalf("Expected CallExpression, got %#v", node)
	}

	if _, ok := expr.Callee.(ComputedAccessExpression); !ok {
		t.Fatalf("Expected callee f[number], got %#v", node)

	}
}

func TestObjectExpression(t *testing.T) {
	tokenizer := testTokenizer{tokens: []Token{
		literal{kind: IDENTIFIER, value: "Type"},
		token{kind: LPAREN},
		literal{kind: IDENTIFIER, value: "value"},
		token{kind: COLON},
		literal{kind: NUMBER, value: "42"},
		token{kind: RPAREN},
	}}
	parser := MakeParser(&tokenizer)
	node := ParseExpression(parser)

	_, ok := node.(CallExpression)
	if !ok {
		t.Fatalf("Expected ObjectExpression, got %#v", node)
	}
	if len(parser.errors) != 0 {
		t.Fatalf("Expected no errors, got %+v: %#v", len(parser.errors), parser.errors)
	}
}

func TestListInstanciation(t *testing.T) {
	tokenizer := testTokenizer{tokens: []Token{
		token{kind: LBRACKET},
		token{kind: RBRACKET},
		token{kind: NUM_KW},
		token{kind: LPAREN},
		literal{kind: NUMBER, value: "1"},
		token{kind: COMMA},
		literal{kind: NUMBER, value: "2"},
		token{kind: RPAREN},
	}}
	parser := MakeParser(&tokenizer)
	node := ParseExpression(parser)

	if len(parser.errors) != 0 {
		t.Fatalf("Expected no errors, got %+v: %#v", len(parser.errors), parser.errors)
	}

	object, ok := node.(CallExpression)
	if !ok {
		t.Fatalf("Expected ObjectExpression, got %#v", node)
	}

	_, ok = object.Callee.(ListTypeExpression)
	if !ok {
		t.Fatalf("Expected a list type, got %#v", object.Callee)
	}
}
