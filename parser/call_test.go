package parser

import (
	"strings"
	"testing"
)

func TestParseMapInstanciation(t *testing.T) {
	parser, err := MakeParser(strings.NewReader("Map(\"key\": \"value\")"))
	if err != nil {
		t.Fatal(err)
	}
	parser.parseAccessExpression()

	if len(parser.errors) > 0 {
		t.Fatalf("Expected no errors, got %#v", parser.errors)
	}
}

func TestCheckImplicitMapInstanciation(t *testing.T) {
	parser, err := MakeParser(strings.NewReader(""))
	if err != nil {
		t.Fatal(err)
	}
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
	if _, ok := alias.Ref.(Map).Key.(String); !ok {
		t.Fatalf("Expected string keys")
	}
}

func TestCheckMapInstanciationMissingTypeArg(t *testing.T) {
	parser, err := MakeParser(strings.NewReader(""))
	if err != nil {
		t.Fatal(err)
	}
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
	parser, err := MakeParser(strings.NewReader(""))
	if err != nil {
		t.Fatal(err)
	}
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
	if _, ok := alias.Ref.(Map).Key.(String); !ok {
		t.Fatalf("Expected string keys")
	}
}

func TestCheckMapEntries(t *testing.T) {
	parser, err := MakeParser(strings.NewReader(""))
	if err != nil {
		t.Fatal(err)
	}
	expr := &CallExpression{
		Callee: &ComputedAccessExpression{
			Expr: &Identifier{Token: literal{kind: Name, value: "Map"}},
			Property: &BracketedExpression{Expr: &TupleExpression{Elements: []Expression{
				&Literal{token{kind: StringKeyword}},
				&Literal{token{kind: StringKeyword}},
			}}},
		},
		Args: &ParenthesizedExpression{Expr: &TupleExpression{Elements: []Expression{
			&Entry{
				Key:   &Literal{literal{kind: StringLiteral, value: "\"a\""}},
				Value: &Literal{literal{kind: StringLiteral, value: "\"value\""}},
			},
			&Entry{
				Key: &BracketedExpression{
					Expr: &Literal{literal{kind: StringLiteral, value: "\"b\""}},
				},
				Value: &Literal{literal{kind: StringLiteral, value: "\"value\""}},
			},
		}}},
	}
	expr.typeCheck(parser)

	if len(parser.errors) > 0 {
		t.Fatalf("Expected no errors, got %#v", parser.errors)
	}
}

func TestCheckMapEntriesBadTypes(t *testing.T) {
	parser, err := MakeParser(strings.NewReader(""))
	if err != nil {
		t.Fatal(err)
	}
	expr := &CallExpression{
		Callee: &ComputedAccessExpression{
			Expr: &Identifier{Token: literal{kind: Name, value: "Map"}},
			Property: &BracketedExpression{Expr: &TupleExpression{Elements: []Expression{
				&Literal{token{kind: StringKeyword}},
				&Literal{token{kind: StringKeyword}},
			}}},
		},
		Args: &ParenthesizedExpression{Expr: &TupleExpression{Elements: []Expression{
			&Entry{
				Key:   &Literal{literal{kind: NumberLiteral, value: "1"}},
				Value: &Literal{literal{kind: StringLiteral, value: "\"value\""}},
			},
			&Entry{
				Key:   &Literal{literal{kind: StringLiteral, value: "\"a\""}},
				Value: &Literal{literal{kind: NumberLiteral, value: "42"}},
			},
		}}},
	}
	expr.typeCheck(parser)

	if len(parser.errors) != 2 {
		t.Fatalf("Expected 2 errors, got %v: %#v", len(parser.errors), parser.errors)
	}
}
