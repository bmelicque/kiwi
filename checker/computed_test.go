package checker

import (
	"testing"

	"github.com/bmelicque/test-parser/parser"
)

func TestGenericWithTypeArgs(t *testing.T) {
	checker := MakeChecker()
	checker.scope.Add("Generic", parser.Loc{}, Type{TypeAlias{
		Name:   "Generic",
		Params: []Generic{{Name: "Type"}},
		Ref:    Object{map[string]ExpressionType{"value": Type{Generic{Name: "Type"}}}},
	}})
	expr := checker.checkComputedAccessExpression(parser.ComputedAccessExpression{
		Expr:     parser.TokenExpression{Token: testToken{parser.IDENTIFIER, "Generic", parser.Loc{}}},
		Property: parser.BracketedExpression{Expr: parser.TokenExpression{Token: testToken{parser.NUM_KW, "number", parser.Loc{}}}},
	})

	if len(checker.errors) != 0 {
		t.Fatalf("Expected no errors, got %#v", checker.errors)
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
		t.Fatalf("Type param should've been set to Primitive{NUMBER}, got %#v", alias.Params[0].Value)
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
		t.Fatalf("Member should've been set to Type{Primitive{NUMBER}}, got %#v", memberType.Value)
	}
}

func TestGenericFunctionWithTypeArgs(t *testing.T) {
	checker := MakeChecker()
	checker.scope.Add("function", parser.Loc{}, Function{
		TypeParams: []Generic{{Name: "Type"}},
		Params:     Tuple{[]ExpressionType{Generic{Name: "Type"}}},
		Returned:   Generic{Name: "Type"},
	})

	expr := checker.checkComputedAccessExpression(parser.ComputedAccessExpression{
		Expr:     parser.TokenExpression{Token: testToken{parser.IDENTIFIER, "function", parser.Loc{}}},
		Property: parser.BracketedExpression{Expr: parser.TokenExpression{Token: testToken{parser.NUM_KW, "number", parser.Loc{}}}},
	})

	if len(checker.errors) != 0 {
		t.Fatalf("Expected no errors, got %#v", checker.errors)
	}

	function, ok := expr.typing.(Function)
	if !ok {
		t.Fatalf("Expected Function type, got %#v", expr.typing)
	}

	if function.Params.elements[0].Kind() != NUMBER {
		t.Fatalf("Param should've been set to Primitive{NUMBER}, got %#v", function.Params.elements[0])
	}

	if function.Returned.Kind() != NUMBER {
		t.Fatalf("Param should've been set to Primitive{NUMBER}, got %#v", function.Returned)
	}
}
