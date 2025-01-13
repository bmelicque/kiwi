package parser

import (
	"reflect"
	"testing"
)

func TestBuildGeneric(t *testing.T) {
	scope := NewScope(ProgramScope)
	scope.Add("Type", Loc{}, Type{Generic{}})
	typing := List{Generic{Name: "Type"}}

	compared := List{Number{}}

	built, ok := typing.build(scope, compared)
	if !ok {
		t.Fatalf("Expected 'ok' to be true (no remaining generics)")
	}

	list, ok := built.(List)
	if !ok {
		t.Fatalf("Expected list type, got %v", reflect.TypeOf(list))
	}

	if _, ok = list.Element.(Number); !ok {
		t.Fatalf("Expected number type, got %v", reflect.TypeOf(list.Element))
	}
}

func TestBuildGenericFromScope(t *testing.T) {
	scope := NewScope(ProgramScope)
	scope.Add("Type", Loc{}, Number{})
	typing := Generic{Name: "Type"}

	built, ok := typing.build(scope, nil)
	if !ok {
		t.Fatalf("Expected 'ok' to be true (no remaining generics)")
	}

	if _, ok = built.(Number); !ok {
		t.Fatalf("Expected number type, got %v", built.Text())
	}
}

func TestBuildTypeAlias(t *testing.T) {
	scope := NewScope(ProgramScope)
	typing := TypeAlias{
		Name:   "Type",
		Params: []Generic{{Name: "Param", Value: Number{}}},
		Ref:    Generic{Name: "Param", Value: Number{}},
	}

	built, ok := typing.build(scope, nil)
	if !ok {
		t.Fatalf("Expected 'ok' to be true (no remaining generics)")
	}

	if _, ok := built.(TypeAlias).Ref.(Number); !ok {
		t.Fatalf("Expected number type, got %#v", built.(TypeAlias).Ref)
	}
}

func TestFunctionExtends(t *testing.T) {
	a := Function{Returned: Number{}}
	b := Function{Returned: Number{}}

	if !a.Extends(b) {
		t.Fatalf("Should've extended!")
	}
}

func TestTrait(t *testing.T) {
	typing := TypeAlias{
		Name: "Type",
		Ref:  newObject(),
		Methods: map[string]ExpressionType{
			"method": Function{Returned: Number{}},
		},
	}
	trait := Trait{
		Self: Generic{Name: "_"},
		Members: map[string]ExpressionType{
			"method": Function{Returned: Number{}},
		},
	}

	if !trait.Extends(typing) {
		t.Fatalf("Should've extended!")
	}
}

// func TestBuildTrait(t *testing.T) {
// 	trait := Trait{
// 		Self: Generic{Name: "Self"},
// 		Members: map[string]ExpressionType{
// 			"method": Function{
// 				Params:   &Tuple{[]ExpressionType{Generic{Name: "Self"}}},
// 				Returned: Generic{Name: "Self"},
// 			},
// 		},
// 	}
// 	s := NewScope(ProgramScope)
// 	s.Add("Self", Loc{}, Type{TypeAlias{Name: "Type", Ref: Object{}}})
// 	built, _ := trait.build(s, nil)

// 	trait, ok := built.(Trait)
// 	if !ok {
// 		t.Fatalf("Trait expected")
// 	}
// 	f, ok := trait.Members["method"].(Function)
// 	if !ok {
// 		t.Fatalf("Function expected")
// 	}
// 	if f.Params.Elements[0].(Type).Value.(TypeAlias).Name != "Type" {
// 		t.Fatalf("'Type' expected")
// 	}
// 	if f.Returned.(Type).Value.(TypeAlias).Name != "Type" {
// 		t.Fatalf("'Type' expected")
// 	}
// }

func TestGetSumTypeMember(t *testing.T) {
	option := makeOptionType(Number{})
	some := option.Ref.(Sum).getMember("Some")
	if _, ok := some.(Number); !ok {
		t.Fatalf("Expected number, got %v", some)
	}
}
