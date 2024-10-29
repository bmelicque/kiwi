package parser

import (
	"testing"
)

func TestParseMapInstanciation(t *testing.T) {
	parser := MakeParser(&testTokenizer{tokens: []Token{
		literal{kind: Name, value: "Map"},
		token{kind: LeftParenthesis},
		literal{kind: StringLiteral, value: "\"key\""},
		token{kind: Colon},
		literal{kind: StringLiteral, value: "\"value\""},
		token{kind: RightParenthesis},
	}})
	parser.parseAccessExpression()

	if len(parser.errors) > 0 {
		t.Fatalf("Expected no errors, got %#v", parser.errors)
	}
}

func TestCheckImplicitMapInstanciation(t *testing.T) {
	parser := MakeParser(nil)
	expr := &CallExpression{
		Callee: &Identifier{Token: literal{kind: Name, value: "Map"}},
		Args: &ParenthesizedExpression{Expr: &TupleExpression{Elements: []Expression{
			&Entry{
				Key:   &Literal{literal{kind: StringLiteral, value: "\"key\""}},
				Value: &Literal{literal{kind: StringLiteral, value: "\"value\""}},
			},
		}}},
	}
	expr.typeCheck(parser)

	if len(parser.errors) > 0 {
		t.Fatalf("Expected no errors, got %#v", parser.errors)
	}
	alias, ok := expr.typing.(TypeAlias)
	if !ok || alias.Name != "Map" {
		t.Fatalf("Map expected")
	}
	if alias.Ref.(Map).Key.Kind() != STRING {
		t.Fatalf("Expected string keys")
	}
}

func TestCheckMapInstanciationMissingTypeArg(t *testing.T) {
	parser := MakeParser(nil)
	expr := &CallExpression{
		Callee: &Identifier{Token: literal{kind: Name, value: "Map"}},
		Args:   &ParenthesizedExpression{Expr: &TupleExpression{Elements: []Expression{}}},
	}
	expr.typeCheck(parser)

	if len(parser.errors) != 1 {
		t.Fatalf("Expected 1 error, got %v: %#v", len(parser.errors), parser.errors)
	}
}

func TestCheckExplicitMapInstanciation(t *testing.T) {
	parser := MakeParser(nil)
	expr := &CallExpression{
		Callee: &ComputedAccessExpression{
			Expr: &Identifier{Token: literal{kind: Name, value: "Map"}},
			Property: &BracketedExpression{Expr: &TupleExpression{Elements: []Expression{
				&Literal{token{kind: StringKeyword}},
				&Literal{token{kind: StringKeyword}},
			}}},
		},
		Args: &ParenthesizedExpression{Expr: &TupleExpression{}},
	}
	expr.typeCheck(parser)

	if len(parser.errors) > 0 {
		t.Fatalf("Expected no errors, got %#v", parser.errors)
	}
	alias, ok := expr.typing.(TypeAlias)
	if !ok || alias.Name != "Map" {
		t.Fatalf("Map expected")
	}
	if alias.Ref.(Map).Key.Kind() != STRING {
		t.Fatalf("Expected string keys")
	}
}
