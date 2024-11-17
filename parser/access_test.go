package parser

import (
	"strings"
	"testing"
)

func TestComputedPropertyAccess(t *testing.T) {
	parser, err := MakeParser(strings.NewReader("n[p]"))
	if err != nil {
		t.Fatal(err)
	}
	node := parser.parseAccessExpression()

	expr, ok := node.(*ComputedAccessExpression)
	if !ok {
		t.Fatalf("Expected ComputedAccessExpression, got %#v", node)
	}
	if _, ok := expr.Expr.(*Identifier); !ok {
		t.Fatalf("Expected token 'n'")
	}
	if _, ok := expr.Property.Expr.(*Identifier); !ok {
		t.Fatalf("Expected token 'p'")
	}
}

func TestPropertyAccess(t *testing.T) {
	parser, err := MakeParser(strings.NewReader("n.p"))
	if err != nil {
		t.Fatal(err)
	}
	node := parser.parseAccessExpression()

	expr, ok := node.(*PropertyAccessExpression)
	if !ok {
		t.Fatalf("Expected PropertyAccessExpression, got %#v", node)
	}
	if _, ok := expr.Expr.(*Identifier); !ok {
		t.Fatalf("Expected token 'n'")
	}
	if _, ok := expr.Property.(*Identifier); !ok {
		t.Fatalf("Expected token 'p'")
	}
}

func TestTupleAccess(t *testing.T) {
	parser, err := MakeParser(strings.NewReader("tuple.0"))
	if err != nil {
		t.Fatal(err)
	}
	parser.scope.Add("tuple", Loc{}, Tuple{[]ExpressionType{Number{}}})
	node := parser.parseAccessExpression()

	expr, ok := node.(*PropertyAccessExpression)
	if !ok {
		t.Fatalf("Expected PropertyAccessExpression, got %#v", node)
	}
	if _, ok := expr.Expr.(*Identifier); !ok {
		t.Fatalf("Expected token 'n'")
	}
	if _, ok := expr.Property.(*Literal); !ok {
		t.Fatalf("Expected literal 0")
	}
}

func TestMethodAccess(t *testing.T) {
	parser, err := MakeParser(strings.NewReader("(t Type).method"))
	if err != nil {
		t.Fatal(err)
	}
	node := parser.parseAccessExpression()

	expr, ok := node.(*PropertyAccessExpression)
	if !ok {
		t.Fatalf("Expected PropertyAccessExpression, got %#v", node)
	}

	if _, ok := expr.Expr.(*ParenthesizedExpression); !ok {
		t.Fatalf("Expected ParenthesizedExpression on LHS, got %#v", expr.Expr)
	}

	if _, ok := expr.Property.(*Identifier); !ok {
		t.Fatalf("Expected token 'method'")
	}
}

func TestTraitDefinition(t *testing.T) {
	parser, err := MakeParser(strings.NewReader("(Self).(method() -> Self)"))
	if err != nil {
		t.Fatal(err)
	}
	node := parser.parseAccessExpression()

	if len(parser.errors) > 0 {
		t.Fatalf("Got %v parsing errors: %#v", len(parser.errors), parser.errors)
	}

	_, ok := node.(*TraitExpression)
	if !ok {
		t.Fatalf("Expected PropertyAccessExpression, got %#v", node)
	}
}

func TestFunctionCall(t *testing.T) {
	parser, err := MakeParser(strings.NewReader("f(42)"))
	if err != nil {
		t.Fatal(err)
	}
	node := parser.parseAccessExpression()

	expr, ok := node.(*CallExpression)
	if !ok {
		t.Fatalf("Expected CallExpression, got %#v", node)
	}
	if _, ok := expr.Callee.(*Identifier); !ok {
		t.Fatalf("Expected token 'f'")
	}
}

func TestFunctionCallWithTypeArgs(t *testing.T) {
	parser, err := MakeParser(strings.NewReader("f[number](42)"))
	if err != nil {
		t.Fatal(err)
	}
	node := parser.parseAccessExpression()

	expr, ok := node.(*CallExpression)
	if !ok {
		t.Fatalf("Expected CallExpression, got %#v", node)
	}

	if _, ok := expr.Callee.(*ComputedAccessExpression); !ok {
		t.Fatalf("Expected callee f[number], got %#v", node)

	}
}

func TestObjectExpression(t *testing.T) {
	parser, err := MakeParser(strings.NewReader("Type(value: 42)"))
	if err != nil {
		t.Fatal(err)
	}
	node := ParseExpression(parser)

	_, ok := node.(*CallExpression)
	if !ok {
		t.Fatalf("Expected CallExpression, got %#v", node)
	}
	if len(parser.errors) != 0 {
		t.Fatalf("Expected no errors, got %+v: %#v", len(parser.errors), parser.errors)
	}
}

func TestListInstanciation(t *testing.T) {
	parser, err := MakeParser(strings.NewReader("[]number(1, 2)"))
	if err != nil {
		t.Fatal(err)
	}
	node := parser.parseExpression()

	if len(parser.errors) != 0 {
		t.Fatalf("Expected no errors, got %+v: %#v", len(parser.errors), parser.errors)
	}

	object, ok := node.(*CallExpression)
	if !ok {
		t.Fatalf("Expected ObjectExpression, got %#v", node)
	}

	_, ok = object.Callee.(*ListTypeExpression)
	if !ok {
		t.Fatalf("Expected a list type, got %#v", object.Callee)
	}
}
