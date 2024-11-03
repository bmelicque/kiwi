package parser

import "testing"

func TestGenericWithTypeArgs(t *testing.T) {
	parser := MakeParser(nil)
	parser.scope.Add("Boxed", Loc{}, Type{TypeAlias{
		Name:   "Boxed",
		Params: []Generic{{Name: "Type"}},
		Ref: Object{map[string]ExpressionType{
			"value": Type{Generic{Name: "Type"}},
		}},
	}})
	// Boxed[number]
	expr := &ComputedAccessExpression{
		Expr:     &Identifier{Token: literal{kind: Name, value: "Boxed"}},
		Property: &BracketedExpression{Expr: &Literal{literal{kind: NumberKeyword}}},
	}
	expr.typeCheck(parser)

	if len(parser.errors) != 0 {
		t.Fatalf("Expected no errors, got %#v", parser.errors)
	}

	typing, ok := expr.typing.(Type)
	if !ok {
		t.Fatalf("Expected Type{TypeAlias{}}, got %#v", expr.typing)
	}

	alias, ok := typing.Value.(TypeAlias)
	if !ok {
		t.Fatalf("Expected Type{TypeAlias{}}, got %#v", typing.Value)
	}

	if alias.Params[0].Value.Kind() != NUMBER {
		t.Fatalf("Type param should've been set to Number{}, got %#v", alias.Params[0].Value)
	}

	object, ok := alias.Ref.(Object)
	if !ok {
		t.Fatalf("Type ref should've been Object{}, got %#v", alias.Ref)
	}

	member, ok := object.Members["value"]
	if !ok {
		t.Fatalf("Could not find member")
	}

	memberType, ok := member.(Type)
	if !ok {
		t.Fatalf("Member should've been a type, got %#v", member)
	}

	if memberType.Value.Kind() != NUMBER {
		t.Fatalf("Member should've been set to Type{Number{}}, got %#v", memberType.Value)
	}
}

func TestGenericFunctionWithTypeArgs(t *testing.T) {
	parser := MakeParser(nil)
	parser.scope.Add("function", Loc{}, Function{
		TypeParams: []Generic{{Name: "Type"}},
		Params:     &Tuple{[]ExpressionType{Generic{Name: "Type"}}},
		Returned:   Generic{Name: "Type"},
	})
	expr := &ComputedAccessExpression{
		Expr:     &Identifier{Token: literal{kind: Name, value: "function"}},
		Property: &BracketedExpression{Expr: &Literal{literal{kind: NumberKeyword}}},
	}
	expr.typeCheck(parser)

	if len(parser.errors) != 0 {
		t.Fatalf("Expected no errors, got %#v", parser.errors)
	}

	function, ok := expr.typing.(Function)
	if !ok {
		t.Fatalf("Expected Function type, got %#v", expr.typing)
	}

	if function.Params.elements[0].Kind() != NUMBER {
		t.Fatalf("Param should've been set to Number{}, got %#v", function.Params.elements[0])
	}

	if function.Returned.Kind() != NUMBER {
		t.Fatalf("Param should've been set to Number{}, got %#v", function.Returned)
	}
}

func TestMapElementAccess(t *testing.T) {
	parser := MakeParser(nil)
	parser.scope.Add("map", Loc{}, makeMapType(Number{}, String{}))
	// map[42]
	expr := &ComputedAccessExpression{
		Expr: &Identifier{Token: literal{kind: Name, value: "map"}},
		Property: &BracketedExpression{
			Expr: &Literal{literal{kind: NumberLiteral, value: "42"}},
		},
	}
	expr.typeCheck(parser)

	if len(parser.errors) != 0 {
		t.Fatalf("Expected no errors, got %#v", parser.errors)
	}

	alias, ok := expr.typing.(TypeAlias)
	if !ok || alias.Name != "Option" {
		t.Fatalf("Expected an option, got %#v", expr.typing)
	}
	some := alias.Ref.(Sum).getMember("Some")
	if some.Kind() != STRING {
		t.Fatalf("Expected option of string, got %#v", some)
	}
}

func TestMapElementAccessBadKey(t *testing.T) {
	parser := MakeParser(nil)
	parser.scope.Add("map", Loc{}, makeMapType(Number{}, String{}))
	// map["42"]
	expr := &ComputedAccessExpression{
		Expr: &Identifier{Token: literal{kind: Name, value: "map"}},
		Property: &BracketedExpression{
			Expr: &Literal{literal{kind: StringLiteral, value: "\"42\""}},
		},
	}
	expr.typeCheck(parser)

	if len(parser.errors) != 1 {
		t.Fatalf("Expected 1 error, got %v: %#v", len(parser.errors), parser.errors)
	}
}
