package parser

import (
	"strings"
	"testing"
)

func TestEmptyBlock(t *testing.T) {
	parser, err := MakeParser(strings.NewReader("{}"))
	if err != nil {
		t.Fatal(err)
	}
	block := parser.parseBlock()

	if len(parser.errors) != 0 {
		t.Fatalf("Expected no errors, got %#v", parser.errors)
	}
	if _, ok := block.Type().(Nil); !ok {
		t.Fatalf("Expected nil type, got %#v", block.Type())
	}
}

func TestSingleLineBlock(t *testing.T) {
	parser, err := MakeParser(strings.NewReader("{ \"Hello, world!\" }"))
	if err != nil {
		t.Fatal(err)
	}
	block := parser.parseBlock()

	if len(parser.errors) != 0 {
		t.Fatalf("Expected no errors, got %#v", parser.errors)
	}
	if _, ok := block.Type().(String); !ok {
		t.Fatalf("Expected string type, got %#v", block.Type())
	}
}

func TestMultilineBlock(t *testing.T) {
	str := "{\n"
	str += "    \"Hello, world!\"\n"
	str += "    \"Hello, world!\"\n"
	str += "}"
	parser, err := MakeParser(strings.NewReader(str))
	if err != nil {
		t.Fatal(err)
	}
	block := parser.parseBlock()

	if len(parser.errors) != 0 {
		t.Fatalf("Expected no errors, got %#v", parser.errors)
	}
	if len(block.Statements) != 2 {
		t.Fatalf("Expected 2 statements, got %#v", block.Statements)
	}
	if _, ok := block.Type().(String); !ok {
		t.Fatalf("Expected string type, got %#v", block.Type())
	}
}

func TestUnreachableCode(t *testing.T) {
	str := "{\n"
	str += "    return \"Hello, world!\"\n"
	str += "    \"Hello, world!\"\n"
	str += "}"
	parser, err := MakeParser(strings.NewReader(str))
	if err != nil {
		t.Fatal(err)
	}
	parser.pushScope(NewScope(FunctionScope))
	parser.parseBlock()

	if len(parser.errors) != 1 {
		t.Fatalf("Expected 1 error, got %#v", parser.errors)
	}
}
