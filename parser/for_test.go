package parser

import (
	"strings"
	"testing"
)

func TestForExpression(t *testing.T) {
	parser, err := MakeParser(strings.NewReader("for true { 42 }"))
	if err != nil {
		t.Fatal(err)
	}
	parser.parseForExpression()

	if len(parser.errors) != 0 {
		t.Fatalf("Expected no errors, got %#v", parser.errors)
	}
}

func TestForExpressionType(t *testing.T) {
	parser, err := MakeParser(strings.NewReader("for true { break 42 }"))
	if err != nil {
		t.Fatal(err)
	}
	expr := parser.parseForExpression()
	expr.typeCheck(parser)

	if len(parser.errors) != 0 {
		t.Fatalf("Expected no errors, got %#v", parser.errors)
	}
	if _, ok := expr.Type().(Number); !ok {
		t.Fatalf("Expected number type, got %#v", expr.Type())
	}
}
