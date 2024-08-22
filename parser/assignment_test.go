package parser

import (
	"testing"

	"github.com/bmelicque/test-parser/tokenizer"
)

func TestAssignment(t *testing.T) {
	tokenizer := testTokenizer{tokens: []tokenizer.Token{
		testToken{tokenizer.IDENTIFIER, "n", tokenizer.Loc{}},
		testToken{tokenizer.ASSIGN, "=", tokenizer.Loc{}},
		testToken{tokenizer.NUMBER, "42", tokenizer.Loc{}},
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
	tokenizer := testTokenizer{tokens: []tokenizer.Token{
		testToken{tokenizer.IDENTIFIER, "n", tokenizer.Loc{}},
		testToken{tokenizer.COMMA, ",", tokenizer.Loc{}},
		testToken{tokenizer.IDENTIFIER, "m", tokenizer.Loc{}},
		testToken{tokenizer.ASSIGN, "=", tokenizer.Loc{}},
		testToken{tokenizer.NUMBER, "1", tokenizer.Loc{}},
		testToken{tokenizer.COMMA, ",", tokenizer.Loc{}},
		testToken{tokenizer.NUMBER, "2", tokenizer.Loc{}},
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
	tokenizer := testTokenizer{tokens: []tokenizer.Token{
		testToken{tokenizer.IDENTIFIER, "Type", tokenizer.Loc{}},
		testToken{tokenizer.DEFINE, "::", tokenizer.Loc{}},
		testToken{tokenizer.LBRACE, "{", tokenizer.Loc{}},
		testToken{tokenizer.EOL, "\n", tokenizer.Loc{}},
		testToken{tokenizer.IDENTIFIER, "n", tokenizer.Loc{}},
		testToken{tokenizer.NUM_KW, "number", tokenizer.Loc{}},
		testToken{tokenizer.COMMA, ",", tokenizer.Loc{}},
		testToken{tokenizer.EOL, "\n", tokenizer.Loc{}},
		testToken{tokenizer.RBRACE, "}", tokenizer.Loc{}},
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
	if _, ok := expr.Initializer.(ObjectDefinition); !ok {
		t.Fatalf("Expected ObjectDefinition")
	}
}

func TestMethodDeclaration(t *testing.T) {
	tokenizer := testTokenizer{tokens: []tokenizer.Token{
		testToken{tokenizer.LPAREN, "(", tokenizer.Loc{}},
		testToken{tokenizer.IDENTIFIER, "t", tokenizer.Loc{}},
		testToken{tokenizer.IDENTIFIER, "Type", tokenizer.Loc{}},
		testToken{tokenizer.RPAREN, ")", tokenizer.Loc{}},
		testToken{tokenizer.DOT, ".", tokenizer.Loc{}},
		testToken{tokenizer.IDENTIFIER, "method", tokenizer.Loc{}},
		testToken{tokenizer.DEFINE, "::", tokenizer.Loc{}},
		testToken{tokenizer.LPAREN, "(", tokenizer.Loc{}},
		testToken{tokenizer.RPAREN, ")", tokenizer.Loc{}},
		testToken{tokenizer.SLIM_ARR, "->", tokenizer.Loc{}},
		testToken{tokenizer.LPAREN, "(", tokenizer.Loc{}},
		testToken{tokenizer.RPAREN, ")", tokenizer.Loc{}},
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
