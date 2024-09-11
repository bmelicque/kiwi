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
