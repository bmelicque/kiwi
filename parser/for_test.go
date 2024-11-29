package parser

import (
	"strings"
	"testing"
)

func TestParseForEmptyExpression(t *testing.T) {
	parser := MakeParser(strings.NewReader("for { 42 }"))
	parser.parseForExpression()

	if len(parser.errors) != 0 {
		t.Fatalf("Expected no errors, got %#v", parser.errors)
	}
}

func TestParseForExpression(t *testing.T) {
	parser := MakeParser(strings.NewReader("for true { 42 }"))
	parser.parseForExpression()

	if len(parser.errors) != 0 {
		t.Fatalf("Expected no errors, got %#v", parser.errors)
	}
}

func TestParseForInExpression(t *testing.T) {
	parser := MakeParser(strings.NewReader("for el in array { 42 }"))
	parser.parseForExpression()

	if len(parser.errors) != 0 {
		t.Fatalf("Expected no errors, got %#v", parser.errors)
	}
}

func TestParseForInTupleExpression(t *testing.T) {
	parser := MakeParser(strings.NewReader("for el, i in array { 42 }"))
	parser.parseForExpression()

	if len(parser.errors) != 0 {
		t.Fatalf("Expected no errors, got %#v", parser.errors)
	}
}

func TestParseForInExpressionMissingIdentifier(t *testing.T) {
	parser := MakeParser(strings.NewReader("for in array { 42 }"))
	parser.parseForExpression()

	if len(parser.errors) != 1 {
		t.Fatalf("Expected 1 error, got %v:\n %#v", len(parser.errors), parser.errors)
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
