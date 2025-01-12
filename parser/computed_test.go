package parser

import (
	"testing"
)

func TestGenericWithTypeArgs(t *testing.T) {
	parser := MakeParser(nil)
	o := newObject()
	o.Members = append(o.Members, ObjectMember{"value", Type{Generic{Name: "Type"}}})
	parser.scope.Add("Boxed", Loc{}, Type{TypeAlias{
		Name:   "Boxed",
		Params: []Generic{{Name: "Type"}},
		Ref:    o,
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

	member, ok := object.GetOwned("value")
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

	if _, ok := function.Params.Elements[0].(Number); !ok {
		t.Fatalf("Param should've been set to Number{}, got %#v", function.Params.Elements[0])
	}

	if _, ok := function.Returned.(Number); !ok {
		t.Fatalf("Param should've been set to Number{}, got %#v", function.Returned)
	}
}
