package parser

import (
	"testing"

	"github.com/bmelicque/test-parser/tokenizer"
)

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
	if expr.Args == nil {
		t.Fatalf("Expected argument 42")
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
	if _, ok := expr.Callee.(TokenExpression); !ok {
		t.Fatalf("Expected token 'f'")
	}
	if expr.Args == nil {
		t.Fatalf("Expected argument 42")
	}
}

func TestGenericTypeCall(t *testing.T) {
	tokenizer := testTokenizer{tokens: []tokenizer.Token{
		testToken{tokenizer.IDENTIFIER, "Type", tokenizer.Loc{}},
		testToken{tokenizer.LBRACKET, "[", tokenizer.Loc{}},
		testToken{tokenizer.IDENTIFIER, "number", tokenizer.Loc{}},
		testToken{tokenizer.RBRACKET, "]", tokenizer.Loc{}},
	}}
	parser := MakeParser(&tokenizer)
	node := parser.parseAccessExpression()

	expr, ok := node.(CallExpression)
	if !ok {
		t.Fatalf("Expected CallExpression, got %#v", node)
	}
	if _, ok := expr.Callee.(TokenExpression); !ok {
		t.Fatalf("Expected token 'Type'")
	}
	if expr.TypeArgs == nil {
		t.Fatalf("Expected type args")
	}
	if expr.Args != nil {
		t.Fatalf("Expected no args")
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

	_, ok := node.(ObjectExpression)
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

	object, ok := node.(ObjectExpression)
	if !ok {
		t.Fatalf("Expected ObjectExpression, got %#v", node)
	}

	_, ok = object.Typing.(ListTypeExpression)
	if !ok {
		t.Fatalf("Expected a list type, got %#v", object.Typing)
	}
}
