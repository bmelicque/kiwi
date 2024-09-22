package checker

import (
	"reflect"
	"testing"

	"github.com/bmelicque/test-parser/tokenizer"
)

func TestBuildGeneric(t *testing.T) {
	scope := NewScope()
	scope.Add("Type", tokenizer.Loc{}, Type{Generic{}})
	typing := List{Generic{Name: "Type"}}

	compared := List{Primitive{NUMBER}}

	built, ok := typing.build(scope, compared)
	if !ok {
		t.Fatalf("Expected 'ok' to be true (no remaining generics)")
	}

	list, ok := built.(List)
	if !ok {
		t.Fatalf("Expected list type, got %v", reflect.TypeOf(list))
	}

	if _, ok = list.Element.(Primitive); !ok {
		t.Fatalf("Expected primitive type, got %v", reflect.TypeOf(list.Element))
	}
}

func TestFunctionExtends(t *testing.T) {
	a := Function{Returned: Primitive{NUMBER}}
	b := Function{Returned: Primitive{NUMBER}}

	if !a.Extends(b) {
		t.Fatalf("Should've extended!")
	}
}

func TestTrait(t *testing.T) {
	typing := TypeAlias{
		Name: "Type",
		Ref:  Object{map[string]ExpressionType{}},
		Methods: map[string]ExpressionType{
			"method": Function{Returned: Primitive{NUMBER}},
		},
	}
	trait := Trait{
		Self: Generic{Name: "_"},
		Members: map[string]ExpressionType{
			"method": Function{Returned: Primitive{NUMBER}},
		},
	}

	if !trait.Extends(typing) {
		t.Fatalf("Should've extended!")
	}
}
