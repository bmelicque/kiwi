package parser

import (
	"strings"
	"testing"
)

func TestEmptyBrackets(t *testing.T) {
	parser, err := MakeParser(strings.NewReader("[]"))
	if err != nil {
		t.Fatal(err)
	}
	parser.parseBracketedExpression()

	if len(parser.errors) != 0 {
		t.Fatalf("Expected no errors, got %v: %#v", len(parser.errors), parser.errors)
	}
}

func TestSimpleBracket(t *testing.T) {
	parser, err := MakeParser(strings.NewReader("[Type]"))
	if err != nil {
		t.Fatal(err)
	}
	parser.parseBracketedExpression()

	if len(parser.errors) != 0 {
		t.Fatalf("Expected no errors, got %v: %#v", len(parser.errors), parser.errors)
	}
}

func TestBracketedTuple(t *testing.T) {
	parser, err := MakeParser(strings.NewReader("[Type, Type]"))
	if err != nil {
		t.Fatal(err)
	}
	parser.parseBracketedExpression()

	if len(parser.errors) != 0 {
		t.Fatalf("Expected no errors, got %v: %#v", len(parser.errors), parser.errors)
	}
}
