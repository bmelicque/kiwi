package parser

import "testing"

func TestSumTypeConstructor1(t *testing.T) {
	parser := MakeParser(nil)
	expr := PropertyAccessExpression{
		Expr:     &Identifier{Token: literal{kind: Name, value: "Option"}, isType: true},
		Property: &Identifier{Token: literal{kind: Name, value: "Some"}, isType: true},
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
	// tuple.1
	parser := MakeParser(&testTokenizer{tokens: []Token{
		literal{kind: Name, value: "tuple"},
		token{kind: Dot},
		literal{kind: NumberLiteral, value: "1"},
	}})
	parser.scope.Add(
		"tuple",
		Loc{},
		Tuple{[]ExpressionType{Primitive{NUMBER}, Primitive{STRING}}},
	)
	expr := parser.parseAccessExpression()
	expr.typeCheck(parser)

	if len(parser.errors) != 0 {
		t.Fatalf("Expected no errors, got %#v", parser.errors)
	}

	if expr.Type().Kind() != STRING {
		t.Fatalf("Expected STRING, got %#v", expr.Type())
	}
}
