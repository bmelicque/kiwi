package parser

import (
	"testing"

	"github.com/bmelicque/test-parser/tokenizer"
)

func TestComputedPropertyAccess(t *testing.T) {
	tokenizer := testTokenizer{tokens: []tokenizer.Token{
		testToken{tokenizer.IDENTIFIER, "n", tokenizer.Loc{}},
		testToken{tokenizer.LBRACKET, "[", tokenizer.Loc{}},
		testToken{tokenizer.IDENTIFIER, "p", tokenizer.Loc{}},
		testToken{tokenizer.RBRACKET, "]", tokenizer.Loc{}},
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
	tokenizer := testTokenizer{tokens: []tokenizer.Token{
		testToken{tokenizer.IDENTIFIER, "n", tokenizer.Loc{}},
		testToken{tokenizer.DOT, ".", tokenizer.Loc{}},
		testToken{tokenizer.IDENTIFIER, "p", tokenizer.Loc{}},
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
	tokenizer := testTokenizer{tokens: []tokenizer.Token{
		testToken{kind: tokenizer.IDENTIFIER, value: "tuple"},
		testToken{kind: tokenizer.DOT},
		testToken{kind: tokenizer.NUMBER, value: "0"},
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
	tokenizer := testTokenizer{tokens: []tokenizer.Token{
		testToken{tokenizer.LPAREN, "(", tokenizer.Loc{}},
		testToken{tokenizer.IDENTIFIER, "t", tokenizer.Loc{}},
		testToken{tokenizer.IDENTIFIER, "Type", tokenizer.Loc{}},
		testToken{tokenizer.RPAREN, ")", tokenizer.Loc{}},
		testToken{tokenizer.DOT, ".", tokenizer.Loc{}},
		testToken{tokenizer.IDENTIFIER, "method", tokenizer.Loc{}},
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
	tokenizer := testTokenizer{tokens: []tokenizer.Token{
		testToken{kind: tokenizer.LPAREN},
		testToken{kind: tokenizer.IDENTIFIER, value: "Self"},
		testToken{kind: tokenizer.RPAREN},
		testToken{kind: tokenizer.DOT},
		testToken{kind: tokenizer.LBRACE},
		testToken{kind: tokenizer.IDENTIFIER, value: "method"},
		testToken{kind: tokenizer.LPAREN},
		testToken{kind: tokenizer.RPAREN},
		testToken{kind: tokenizer.SLIM_ARR},
		testToken{kind: tokenizer.IDENTIFIER, value: "Self"},
		testToken{kind: tokenizer.RBRACE},
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

	if _, ok := expr.Property.(ObjectDefinition); !ok {
		t.Fatalf("Expected ObjectDefinition, got %#v", expr.Property)
	}
}

func TestFunctionCall(t *testing.T) {
	tokenizer := testTokenizer{tokens: []tokenizer.Token{
		testToken{tokenizer.IDENTIFIER, "f", tokenizer.Loc{}},
		testToken{tokenizer.LPAREN, "(", tokenizer.Loc{}},
		testToken{tokenizer.NUMBER, "42", tokenizer.Loc{}},
		testToken{tokenizer.RPAREN, ")", tokenizer.Loc{}},
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
	tokenizer := testTokenizer{tokens: []tokenizer.Token{
		testToken{tokenizer.IDENTIFIER, "f", tokenizer.Loc{}},
		testToken{tokenizer.LBRACKET, "[", tokenizer.Loc{}},
		testToken{tokenizer.IDENTIFIER, "number", tokenizer.Loc{}},
		testToken{tokenizer.RBRACKET, "]", tokenizer.Loc{}},
		testToken{tokenizer.LPAREN, "(", tokenizer.Loc{}},
		testToken{tokenizer.NUMBER, "42", tokenizer.Loc{}},
		testToken{tokenizer.RPAREN, ")", tokenizer.Loc{}},
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
	tokenizer := testTokenizer{tokens: []tokenizer.Token{
		testToken{tokenizer.IDENTIFIER, "Type", tokenizer.Loc{}},
		testToken{tokenizer.LBRACE, "{", tokenizer.Loc{}},
		testToken{tokenizer.IDENTIFIER, "value", tokenizer.Loc{}},
		testToken{tokenizer.COLON, ":", tokenizer.Loc{}},
		testToken{tokenizer.NUMBER, "42", tokenizer.Loc{}},
		testToken{tokenizer.RBRACE, "}", tokenizer.Loc{}},
	}}
	parser := MakeParser(&tokenizer)
	node := ParseExpression(parser)

	_, ok := node.(InstanciationExpression)
	if !ok {
		t.Fatalf("Expected ObjectExpression, got %#v", node)
	}
	if len(parser.errors) != 0 {
		t.Fatalf("Expected no errors, got %+v: %#v", len(parser.errors), parser.errors)
	}
}

func TestListInstanciation(t *testing.T) {
	tokenizer := testTokenizer{tokens: []tokenizer.Token{
		testToken{tokenizer.LBRACKET, "[", tokenizer.Loc{}},
		testToken{tokenizer.RBRACKET, "]", tokenizer.Loc{}},
		testToken{tokenizer.NUM_KW, "number", tokenizer.Loc{}},
		testToken{tokenizer.LBRACE, "{", tokenizer.Loc{}},
		testToken{tokenizer.NUMBER, "1", tokenizer.Loc{}},
		testToken{tokenizer.COMMA, ",", tokenizer.Loc{}},
		testToken{tokenizer.NUMBER, "2", tokenizer.Loc{}},
		testToken{tokenizer.RBRACE, "}", tokenizer.Loc{}},
	}}
	parser := MakeParser(&tokenizer)
	node := ParseExpression(parser)

	if len(parser.errors) != 0 {
		t.Fatalf("Expected no errors, got %+v: %#v", len(parser.errors), parser.errors)
	}

	object, ok := node.(InstanciationExpression)
	if !ok {
		t.Fatalf("Expected ObjectExpression, got %#v", node)
	}

	_, ok = object.Typing.(ListTypeExpression)
	if !ok {
		t.Fatalf("Expected a list type, got %#v", object.Typing)
	}
}
