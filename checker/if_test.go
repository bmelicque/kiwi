package checker

import (
	"testing"

	"github.com/bmelicque/test-parser/parser"
	"github.com/bmelicque/test-parser/tokenizer"
)

func TestIfStatement(t *testing.T) {
	checker := MakeChecker()
	expr := checker.checkIf(parser.IfElse{
		Keyword:   testToken{kind: tokenizer.IF_KW},
		Condition: parser.TokenExpression{Token: testToken{kind: tokenizer.BOOLEAN, value: "true"}},
		Body:      &parser.Block{Statements: []parser.Node{}},
	})

	if len(checker.errors) != 0 {
		t.Fatalf("Expected no errors, got %#v", checker.errors)
	}
	_ = expr
}

func TestIfStatementWithNonBoolean(t *testing.T) {
	checker := MakeChecker()
	expr := checker.checkIf(parser.IfElse{
		Keyword:   testToken{kind: tokenizer.IF_KW},
		Condition: parser.TokenExpression{Token: testToken{kind: tokenizer.NUMBER, value: "42"}},
		Body:      &parser.Block{Statements: []parser.Node{}},
	})

	if len(checker.errors) != 1 {
		t.Fatalf("Expected one error, got %#v", checker.errors)
	}
	_ = expr
}

func TestIfElseStatements(t *testing.T) {
	checker := MakeChecker()
	expr := checker.checkIf(parser.IfElse{
		Keyword:   testToken{kind: tokenizer.IF_KW},
		Condition: parser.TokenExpression{Token: testToken{kind: tokenizer.BOOLEAN, value: "true"}},
		Body:      &parser.Block{Statements: []parser.Node{}},
		Alternate: parser.Block{Statements: []parser.Node{}},
	})

	if len(checker.errors) != 0 {
		t.Fatalf("Expected no errors, got %#v", checker.errors)
	}
	if expr.Alternate == nil {
		t.Fatalf("Expected alternate")
	}
	if _, ok := expr.Alternate.(Body); !ok {
		t.Fatalf("Expected body as alternate")
	}
}

func TestIfElseIfStatements(t *testing.T) {
	checker := MakeChecker()
	// if false {} else if true {}
	expr := checker.checkIf(parser.IfElse{
		Keyword:   testToken{kind: tokenizer.IF_KW},
		Condition: parser.TokenExpression{Token: testToken{kind: tokenizer.BOOLEAN, value: "false"}},
		Body:      &parser.Block{Statements: []parser.Node{}},
		Alternate: parser.IfElse{
			Condition: parser.TokenExpression{Token: testToken{kind: tokenizer.BOOLEAN, value: "true"}},
			Body:      &parser.Block{Statements: []parser.Node{}},
		},
	})

	if len(checker.errors) != 0 {
		t.Fatalf("Expected no errors, got %#v", checker.errors)
	}
	if expr.Alternate == nil {
		t.Fatalf("Expected alternate")
	}
	if _, ok := expr.Alternate.(If); !ok {
		t.Fatalf("Expected body as alternate")
	}
}
