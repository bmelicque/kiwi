package parser

import (
	"strings"
	"testing"
)

func TestForExpression(t *testing.T) {
	parser := MakeParser(strings.NewReader("for true { 42 }"))
	parser.parseForExpression()

	if len(parser.errors) != 0 {
		t.Fatalf("Expected no errors, got %#v", parser.errors)
	}
}

func TestForExpressionType(t *testing.T) {
	parser := MakeParser(strings.NewReader("for true { break 42 }"))
	expr := parser.parseForExpression()
	expr.typeCheck(parser)

	if len(parser.errors) != 0 {
		t.Fatalf("Expected no errors, got %#v", parser.errors)
	}
	if _, ok := expr.Type().(Number); !ok {
		t.Fatalf("Expected number type, got %#v", expr.Type())
	}
}
