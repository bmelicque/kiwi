package parser

import (
	"strings"
	"testing"
)

func TestCheckPropertyAccess(t *testing.T) {
	parser, _ := MakeParser(nil)
	alias := TypeAlias{
		Name: "BoxedNumber",
		Ref: Object{Members: map[string]ExpressionType{
			"value": Number{},
		}},
	}
	parser.scope.Add("BoxedNumber", Loc{}, Type{alias})
	parser.scope.Add("box", Loc{}, alias)
	expr := PropertyAccessExpression{
		Expr:     &Identifier{Token: literal{kind: Name, value: "box"}},
		Property: &Identifier{Token: literal{kind: Name, value: "value"}},
	}
	expr.typeCheck(parser)

	if len(parser.errors) != 0 {
		t.Fatalf("Expected no errors, got %#v", parser.errors)
	}

	if _, ok := expr.Type().(Number); !ok {
		t.Fatalf("Expected number, got %v", expr.Type().Text())
	}
}

func TestCheckPropertyAccessThroughRef(t *testing.T) {
	parser, _ := MakeParser(nil)
	alias := TypeAlias{
		Name: "BoxedNumber",
		Ref: Object{Members: map[string]ExpressionType{
			"value": Number{},
		}},
	}
	parser.scope.Add("BoxedNumber", Loc{}, Type{alias})
	parser.scope.Add("ref", Loc{}, Ref{alias})
	expr := PropertyAccessExpression{
		Expr:     &Identifier{Token: literal{kind: Name, value: "ref"}},
		Property: &Identifier{Token: literal{kind: Name, value: "value"}},
	}
	expr.typeCheck(parser)

	if len(parser.errors) != 0 {
		t.Fatalf("Expected no errors, got %#v", parser.errors)
	}

	if _, ok := expr.Type().(Number); !ok {
		t.Fatalf("Expected number, got %v", expr.Type().Text())
	}
}

func TestSumTypeConstructor1(t *testing.T) {
	parser, _ := MakeParser(nil)
	expr := PropertyAccessExpression{
		Expr:     &Identifier{Token: literal{kind: Name, value: "?"}},
		Property: &Identifier{Token: literal{kind: Name, value: "Some"}},
	}
	expr.typeCheck(parser)

	if len(parser.errors) != 0 {
		t.Fatalf("Expected no errors, got %#v", parser.errors)
	}

	constructor, ok := expr.typing.(Function)
	if !ok {
		t.Fatalf("Expected function, got %#v", expr.typing)
	}

	ret, ok := constructor.Returned.(TypeAlias)
	if !ok {
		t.Fatalf("Expected return to be an alias, got %#v", constructor.Returned)
	}
	if _, ok = ret.Ref.(Sum); !ok {
		t.Fatalf("Expected sum, got %#v", constructor)
	}
}

func TestTupleIndexAccess(t *testing.T) {
	parser, err := MakeParser(strings.NewReader("tuple.1"))
	if err != nil {
		t.Fatal(err)
	}
	parser.scope.Add(
		"tuple",
		Loc{},
		Tuple{[]ExpressionType{Number{}, String{}}},
	)
	expr := parser.parseAccessExpression()
	expr.typeCheck(parser)

	if len(parser.errors) != 0 {
		t.Fatalf("Expected no errors, got %#v", parser.errors)
	}

	if _, ok := expr.Type().(String); !ok {
		t.Fatalf("Expected string, got %#v", expr.Type())
	}
}
