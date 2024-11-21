package parser

import (
	"strings"
	"testing"
)

func TestEmptyBrackets(t *testing.T) {
	parser := MakeParser(strings.NewReader("[]"))
	parser.parseBracketedExpression()

	if len(parser.errors) != 0 {
		t.Fatalf("Expected no errors, got %v: %#v", len(parser.errors), parser.errors)
	}
}

func TestSimpleBracket(t *testing.T) {
	parser := MakeParser(strings.NewReader("[Type]"))
	parser.parseBracketedExpression()

	if len(parser.errors) != 0 {
		t.Fatalf("Expected no errors, got %v: %#v", len(parser.errors), parser.errors)
	}
}

func TestBracketedTuple(t *testing.T) {
	parser := MakeParser(strings.NewReader("[Type, Type]"))
	parser.parseBracketedExpression()

	if len(parser.errors) != 0 {
		t.Fatalf("Expected no errors, got %v: %#v", len(parser.errors), parser.errors)
	}
}
