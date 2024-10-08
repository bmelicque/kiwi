package checker

import (
	"testing"

	"github.com/bmelicque/test-parser/parser"
)

func TestIfExpression(t *testing.T) {
	checker := MakeChecker()
	expr := checker.checkIf(parser.IfElse{
		Keyword:   testToken{kind: parser.IfKeyword},
		Condition: parser.TokenExpression{Token: testToken{kind: parser.BooleanLiteral, value: "true"}},
		Body: &parser.Block{Statements: []parser.Node{
			parser.TokenExpression{Token: testToken{kind: parser.NumberLiteral, value: "42"}},
		}},
	})

	if len(checker.errors) != 0 {
		t.Fatalf("Expected no errors, got %#v", checker.errors)
	}
	if expr.Type().Kind() != NUMBER {
		t.Fatalf("Expected number type")
	}
}

func TestIfExpressionWithNonBoolean(t *testing.T) {
	checker := MakeChecker()
	expr := checker.checkIf(parser.IfElse{
		Keyword:   testToken{kind: parser.IfKeyword},
		Condition: parser.TokenExpression{Token: testToken{kind: parser.NumberLiteral, value: "42"}},
		Body:      &parser.Block{Statements: []parser.Node{}},
	})

	if len(checker.errors) != 1 {
		t.Fatalf("Expected one error, got %#v", checker.errors)
	}
	_ = expr
}

func TestIfElseExpression(t *testing.T) {
	checker := MakeChecker()
	expr := checker.checkIf(parser.IfElse{
		Keyword:   testToken{kind: parser.IfKeyword},
		Condition: parser.TokenExpression{Token: testToken{kind: parser.BooleanLiteral, value: "true"}},
		Body:      &parser.Block{Statements: []parser.Node{}},
		Alternate: parser.Block{Statements: []parser.Node{}},
	})

	if len(checker.errors) != 0 {
		t.Fatalf("Expected no errors, got %#v", checker.errors)
	}
	if expr.Alternate == nil {
		t.Fatalf("Expected alternate")
	}
	if _, ok := expr.Alternate.(Block); !ok {
		t.Fatalf("Expected body as alternate")
	}
}

func TestIfElseExpressionTypeMismatch(t *testing.T) {
	checker := MakeChecker()
	// if true { 42 } else { "Hello, world!" }
	checker.checkIf(parser.IfElse{
		Keyword:   testToken{kind: parser.IfKeyword},
		Condition: parser.TokenExpression{Token: testToken{kind: parser.BooleanLiteral, value: "true"}},
		Body: &parser.Block{Statements: []parser.Node{
			parser.TokenExpression{Token: testToken{kind: parser.NumberLiteral, value: "42"}},
		}},
		Alternate: parser.Block{Statements: []parser.Node{
			parser.TokenExpression{Token: testToken{kind: parser.StringLiteral, value: "\"Hello, world!\""}},
		}},
	})

	if len(checker.errors) != 1 {
		t.Fatalf("Expected 1 error, got %#v", checker.errors)
	}
}

func TestIfElseIfExpression(t *testing.T) {
	checker := MakeChecker()
	// if false {} else if true {}
	expr := checker.checkIf(parser.IfElse{
		Keyword:   testToken{kind: parser.IfKeyword},
		Condition: parser.TokenExpression{Token: testToken{kind: parser.BooleanLiteral, value: "false"}},
		Body:      &parser.Block{Statements: []parser.Node{}},
		Alternate: parser.IfElse{
			Condition: parser.TokenExpression{Token: testToken{kind: parser.BooleanLiteral, value: "true"}},
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
