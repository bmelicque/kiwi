package parser

import (
	"strings"
	"testing"
)

func TestParseLiteral(t *testing.T) {
	parser, err := MakeParser(strings.NewReader("true"))
	if err != nil {
		t.Fatal(err)
	}
	expr := parser.parseToken()
	if len(parser.errors) > 0 {
		t.Fatalf("Expected no errors, got %#v", parser.errors)
	}
	if _, ok := expr.(*Literal); !ok {
		t.Fatalf("Expected Literal, got %#v", expr)
	}
	if _, ok := expr.Type().(Boolean); !ok {
		t.Fatalf("Expected boolean, got %#v", expr.Type())
	}
}

func TestParseIdentifier(t *testing.T) {
	parser, err := MakeParser(strings.NewReader("myVariable"))
	if err != nil {
		t.Fatal(err)
	}
	expr := parser.parseToken()
	if len(parser.errors) > 0 {
		t.Fatalf("Expected no errors, got %#v", parser.errors)
	}
	if _, ok := expr.(*Identifier); !ok {
		t.Fatalf("Expected Identifier, got %#v", expr)
	}
}

func TestCheckIdentifier(t *testing.T) {
	parser, err := MakeParser(strings.NewReader(""))
	if err != nil {
		t.Fatal(err)
	}
	parser.scope.Add("myVariable", Loc{}, Boolean{})
	expr := &Identifier{Token: literal{kind: Name, value: "myVariable"}}
	expr.typeCheck(parser)
	if len(parser.errors) > 0 {
		t.Fatalf("Expected no errors, got %#v", parser.errors)
	}
	if _, ok := expr.Type().(Boolean); !ok {
		t.Fatalf("Expected boolean, got %#v", expr.Type())
	}
}

func TestParseCatchall(t *testing.T) {
	parser, err := MakeParser(strings.NewReader("_"))
	if err != nil {
		t.Fatal(err)
	}
	expr := parser.parseToken()
	if len(parser.errors) > 0 {
		t.Fatalf("Expected no errors, got %#v", parser.errors)
	}
	if _, ok := expr.(*Identifier); !ok {
		t.Fatalf("Expected Identifier, got %#v", expr)
	}
}
