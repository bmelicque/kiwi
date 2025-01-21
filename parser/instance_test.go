package parser

import (
	"strings"
	"testing"
)

func TestParseExplicitGenericInstanciation(t *testing.T) {
	source := "Boxed[number]{ value: 42 }"
	parser := MakeParser(strings.NewReader(source))
	expr := parser.parseStatement()
	testParserErrors(t, parser, 0)
	i, ok := expr.(*InstanceExpression)
	if !ok {
		t.Fatal("Expected *InstanceExpression")
	}
	computed, ok := i.Typing.(*ComputedAccessExpression)
	if !ok {
		t.Fatal("Expected *ComputedAccessExpression")
	}
	if id, ok := computed.Expr.(*Identifier); !ok || id.Text() != "Boxed" {
		t.Fatal("Expected 'Boxed'")
	}
	if _, ok := computed.Property.Expr.(*Literal); !ok {
		t.Fatalf("Expected *Literal, got %#v", computed.Property.Expr)
	}
}

func TestCheckExplicitGenericInstanciation(t *testing.T) {
	parser := MakeParser(nil)
	parser.scope.Add("Boxed", Loc{}, Type{TypeAlias{
		Name:   "Boxed",
		Params: []Generic{{Name: "Type"}},
		Ref:    Object{Members: []ObjectMember{{Name: "value", Type: Generic{Name: "Type"}}}},
	}})
	expr := &InstanceExpression{
		Typing: &ComputedAccessExpression{
			Expr:     &Identifier{Token: literal{kind: Name, value: "Boxed"}},
			Property: &BracketedExpression{Expr: &Literal{token{kind: NumberKeyword}}},
		},
		Args: &BracedExpression{Expr: &TupleExpression{Elements: []Expression{
			&Entry{
				Key:   &Identifier{Token: literal{kind: Name, value: "value"}},
				Value: &Literal{literal{kind: NumberLiteral, value: "42"}},
			},
		}}},
	}
	expr.typeCheck(parser)

	testParserErrors(t, parser, 0)
}

func TestParseMultilineInstanciation(t *testing.T) {
	source := "Type{\n"
	source += "    key: value,\n"
	source += "}\n"
	parser := MakeParser(strings.NewReader(source))
	parser.parseInstanceExpression()

	if len(parser.errors) > 0 {
		t.Logf("Expected no errors, got:")
		for _, err := range parser.errors {
			t.Logf("%v\n", err.Text())
		}
		t.Fail()
	}
}

func TestParseMapInstanciation(t *testing.T) {
	parser := MakeParser(strings.NewReader("Map{\"key\": \"value\"}"))
	parser.parseInstanceExpression()

	if len(parser.errors) > 0 {
		t.Fatalf("Expected no errors, got %#v", parser.errors)
	}
}

func TestCheckImplicitMapInstanciation(t *testing.T) {
	parser := MakeParser(nil)
	expr := &InstanceExpression{
		Typing: &Identifier{Token: literal{kind: Name, value: "Map"}},
		Args: &BracedExpression{Expr: &TupleExpression{Elements: []Expression{
			&Entry{
				Key:   &Literal{literal{kind: StringLiteral, value: "\"key\""}},
				Value: &Literal{literal{kind: StringLiteral, value: "\"value\""}},
			},
		},
		}},
	}
	expr.typeCheck(parser)

	if len(parser.errors) > 0 {
		t.Fatalf("Expected no errors, got %#v", parser.errors)
	}
	alias, ok := expr.Type().(TypeAlias)
	if !ok || alias.Name != "Map" {
		t.Fatalf("Map expected")
	}
	if _, ok := alias.Ref.(Map).Key.(String); !ok {
		t.Fatalf("Expected string keys, got %v", alias.Ref.(Map).Key.Text())
	}
}

func TestCheckMapInstanciationMissingTypeArg(t *testing.T) {
	parser := MakeParser(nil)
	expr := &InstanceExpression{
		Typing: &Identifier{Token: literal{kind: Name, value: "Map"}},
		Args:   &BracedExpression{Expr: makeTuple(nil)},
	}
	expr.typeCheck(parser)

	if len(parser.errors) != 1 {
		t.Fatalf("Expected 1 error, got %v: %#v", len(parser.errors), parser.errors)
	}
}

func TestCheckExplicitMapInstanciation(t *testing.T) {
	parser := MakeParser(nil)
	expr := &InstanceExpression{
		Typing: &ComputedAccessExpression{
			Expr: &Identifier{Token: literal{kind: Name, value: "Map"}},
			Property: &BracketedExpression{Expr: &TupleExpression{Elements: []Expression{
				&Literal{token{kind: StringKeyword}},
				&Literal{token{kind: StringKeyword}},
			}}},
		},
		Args: &BracedExpression{Expr: makeTuple(nil)},
	}
	expr.typeCheck(parser)

	if len(parser.errors) > 0 {
		t.Fatalf("Expected no errors, got %#v", parser.errors)
	}
	alias, ok := expr.Type().(TypeAlias)
	if !ok || alias.Name != "Map" {
		t.Fatalf("Map expected")
	}
	if _, ok := alias.Ref.(Map).Key.(String); !ok {
		t.Fatalf("Expected string keys")
	}
}

func TestCheckMapEntries(t *testing.T) {
	parser := MakeParser(nil)
	expr := &InstanceExpression{
		Typing: &ComputedAccessExpression{
			Expr: &Identifier{Token: literal{kind: Name, value: "Map"}},
			Property: &BracketedExpression{Expr: &TupleExpression{Elements: []Expression{
				&Literal{token{kind: StringKeyword}},
				&Literal{token{kind: StringKeyword}},
			}}},
		},
		Args: &BracedExpression{Expr: &TupleExpression{Elements: []Expression{
			&Entry{
				Key:   &Literal{literal{kind: StringLiteral, value: "\"a\""}},
				Value: &Literal{literal{kind: StringLiteral, value: "\"value\""}},
			},
			&Entry{
				Key:   &Literal{literal{kind: StringLiteral, value: "\"b\""}},
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
	parser := MakeParser(nil)
	expr := &InstanceExpression{
		Typing: &ComputedAccessExpression{
			Expr: &Identifier{Token: literal{kind: Name, value: "Map"}},
			Property: &BracketedExpression{Expr: &TupleExpression{Elements: []Expression{
				&Literal{token{kind: StringKeyword}},
				&Literal{token{kind: StringKeyword}},
			}}},
		},
		Args: &BracedExpression{Expr: &TupleExpression{Elements: []Expression{
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

func TestListTypeInstance(t *testing.T) {
	parser := MakeParser(strings.NewReader("[]number{0, 1, 2}"))
	node := parser.parseExpression()

	if len(parser.errors) != 0 {
		t.Fatalf("Expected no errors, got %+v: %#v", len(parser.errors), parser.errors)
	}

	if _, ok := node.(*InstanceExpression); !ok {
		t.Fatalf("Expected InstanceExpression, got %#v", node)
	}
}

func TestParseAnonymousList(t *testing.T) {
	parser := MakeParser(strings.NewReader("[]{1, 2, 3}"))
	expr := parser.parseInstanceExpression()

	if len(parser.errors) > 0 {
		t.Fatalf("Expected no errors, got %#v", parser.errors)
	}

	i, ok := expr.(*InstanceExpression)
	if !ok {
		t.Fatalf("Expected *InstanceExpression")
	}
	if _, ok := i.Typing.(*ListTypeExpression); !ok {
		t.Fatalf("Expected *ListTypeExpression, got %#v", i.Typing)
	}
}

func TestCheckAnonymousList(t *testing.T) {
	parser := MakeParser(nil)
	expr := &InstanceExpression{
		Typing: &ListTypeExpression{
			Bracketed: &BracketedExpression{},
		},
		Args: &BracedExpression{Expr: &TupleExpression{Elements: []Expression{
			&Literal{literal{kind: NumberLiteral, value: "42"}},
		}}},
	}
	expr.typeCheck(parser)

	if len(parser.errors) > 0 {
		t.Fatalf("Expected no errors, got %#v", parser.errors)
	}
	_, ok := expr.Type().(List)
	if !ok {
		t.Fatalf("List expected")
	}
}
