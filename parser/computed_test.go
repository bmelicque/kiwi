package parser

import (
	"strings"
	"testing"
)

func TestGenericWithTypeArgs(t *testing.T) {
	parser, err := MakeParser(strings.NewReader(""))
	if err != nil {
		t.Fatal(err)
	}
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

	if _, ok := alias.Params[0].Value.(Number); !ok {
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

	if _, ok := memberType.Value.(Number); !ok {
		t.Fatalf("Member should've been set to Type{Number{}}, got %#v", memberType.Value)
	}
}

func TestGenericFunctionWithTypeArgs(t *testing.T) {
	parser, err := MakeParser(strings.NewReader(""))
	if err != nil {
		t.Fatal(err)
	}
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

	if _, ok := function.Params.elements[0].(Number); !ok {
		t.Fatalf("Param should've been set to Number{}, got %#v", function.Params.elements[0])
	}

	if _, ok := function.Returned.(Number); !ok {
		t.Fatalf("Param should've been set to Number{}, got %#v", function.Returned)
	}
}

func TestMapElementAccess(t *testing.T) {
	parser, err := MakeParser(strings.NewReader(""))
	if err != nil {
		t.Fatal(err)
	}
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
	if !ok || alias.Name != "?" {
		t.Fatalf("Expected an option, got %#v", expr.typing)
	}
	some := alias.Ref.(Sum).getMember("Some")
	if _, ok := some.(String); !ok {
		t.Fatalf("Expected option of string, got %#v", some)
	}
}

func TestMapElementAccessBadKey(t *testing.T) {
	parser, err := MakeParser(strings.NewReader(""))
	if err != nil {
		t.Fatal(err)
	}
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

func TestCheckListElementAccess(t *testing.T) {
	parser, err := MakeParser(strings.NewReader(""))
	if err != nil {
		t.Fatal(err)
	}
	parser.scope.Add("list", Loc{}, List{String{}})
	// list[42]
	expr := &ComputedAccessExpression{
		Expr: &Identifier{Token: literal{kind: Name, value: "list"}},
		Property: &BracketedExpression{
			Expr: &Literal{literal{kind: NumberLiteral, value: "42"}},
		},
	}
	expr.typeCheck(parser)

	if len(parser.errors) != 0 {
		t.Fatalf("Expected no errors, got %#v", parser.errors)
	}

	alias, ok := expr.typing.(TypeAlias)
	if !ok || alias.Name != "?" {
		t.Fatalf("Expected an option, got %#v", expr.typing.Text())
	}
	some := alias.Ref.(Sum).getMember("Some")
	if _, ok := some.(String); !ok {
		t.Fatalf("Expected option of string, got %#v", some.Text())
	}
}

func TestCheckListSlice(t *testing.T) {
	parser, err := MakeParser(strings.NewReader(""))
	if err != nil {
		t.Fatal(err)
	}
	parser.scope.Add("list", Loc{}, List{String{}})
	// list[1..]
	expr := &ComputedAccessExpression{
		Expr: &Identifier{Token: literal{kind: Name, value: "list"}},
		Property: &BracketedExpression{
			Expr: &RangeExpression{
				Left:     &Literal{literal{kind: NumberLiteral, value: "1"}},
				Operator: token{kind: ExclusiveRange},
			},
		},
	}
	expr.typeCheck(parser)

	if len(parser.errors) != 0 {
		t.Fatalf("Expected no errors, got %#v", parser.errors)
	}

	list, ok := expr.typing.(List)
	if !ok {
		t.Fatalf("Expected []string, got %#v", expr.typing.Text())
	}
	if _, ok := list.Element.(String); !ok {
		t.Fatalf("Expected option of string, got %#v", expr.typing.Text())
	}
}

func TestCheckListBadIndexType(t *testing.T) {
	parser, err := MakeParser(strings.NewReader(""))
	if err != nil {
		t.Fatal(err)
	}
	parser.scope.Add("list", Loc{}, List{String{}})
	// list["42"]
	expr := &ComputedAccessExpression{
		Expr: &Identifier{Token: literal{kind: Name, value: "list"}},
		Property: &BracketedExpression{
			Expr: &Literal{literal{kind: StringLiteral, value: "\"42\""}},
		},
	}
	expr.typeCheck(parser)

	if len(parser.errors) != 1 {
		t.Fatalf("Expected 1 error, got %#v", parser.errors)
	}

	_, ok := expr.typing.(Unknown)
	if !ok {
		t.Fatalf("Expected unknown, got %#v", expr.typing.Text())
	}
}
