package parser

import "testing"

func TestParseLiteral(t *testing.T) {
	parser := MakeParser(&testTokenizer{tokens: []Token{
		literal{kind: BooleanLiteral, value: "true"},
	}})
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
	parser := MakeParser(&testTokenizer{tokens: []Token{
		literal{kind: Name, value: "myVariable"},
	}})
	expr := parser.parseToken()
	if len(parser.errors) > 0 {
		t.Fatalf("Expected no errors, got %#v", parser.errors)
	}
	if _, ok := expr.(*Identifier); !ok {
		t.Fatalf("Expected Identifier, got %#v", expr)
	}
}

func TestCheckIdentifier(t *testing.T) {
	parser := MakeParser(nil)
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
	parser := MakeParser(&testTokenizer{tokens: []Token{
		literal{kind: Name, value: "_"},
	}})
	expr := parser.parseToken()
	if len(parser.errors) > 0 {
		t.Fatalf("Expected no errors, got %#v", parser.errors)
	}
	if _, ok := expr.(*Identifier); !ok {
		t.Fatalf("Expected Identifier, got %#v", expr)
	}
}
