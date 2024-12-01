package parser

import (
	"strings"
	"testing"
)

func TestAssignment(t *testing.T) {
	parser := MakeParser(strings.NewReader("n = 42"))
	node := parser.parseAssignment()

	if len(parser.errors) != 0 {
		t.Fatalf("Expected no errors, got %#v", parser.errors)
	}

	expr, ok := node.(*Assignment)
	if !ok {
		t.Fatalf("Expected Assignment, got %#v", node)
	}
	if _, ok := expr.Pattern.(*Identifier); !ok {
		t.Fatalf("Expected token 'n'")
	}
	if _, ok := expr.Value.(*Literal); !ok {
		t.Fatalf("Expected literal 42")
	}
}

func TestCheckAssignmentToIdentifier(t *testing.T) {
	parser := MakeParser(nil)
	parser.scope.Add("value", Loc{}, Number{})
	assignment := &Assignment{
		Pattern:  &Identifier{Token: literal{kind: Name, value: "value"}},
		Value:    &Literal{literal{kind: NumberLiteral, value: "42"}},
		Operator: token{kind: Assign},
	}
	assignment.typeCheck(parser)
	if len(parser.errors) != 0 {
		t.Fatalf("Expected no errors, got %#v", parser.errors)
	}
}

func TestCheckAssignmentToIdentifierBadType(t *testing.T) {
	parser := MakeParser(nil)
	parser.scope.Add("value", Loc{}, Number{})
	assignment := &Assignment{
		Pattern:  &Identifier{Token: literal{kind: Name, value: "value"}},
		Value:    &Literal{literal{kind: StringLiteral, value: "\"Hi!\""}},
		Operator: token{kind: Assign},
	}
	assignment.typeCheck(parser)
	if len(parser.errors) != 1 {
		t.Fatalf("Expected 1 error, got %#v", parser.errors)
	}
}

func TestTupleAssignment(t *testing.T) {
	parser := MakeParser(strings.NewReader("n, m = 1, 2"))
	node := parser.parseAssignment()

	expr, ok := node.(*Assignment)
	if !ok {
		t.Fatalf("Expected Assignment, got %#v", node)
	}
	if _, ok := expr.Pattern.(*TupleExpression); !ok {
		t.Fatalf("Expected tuple 'n, m'")
	}
	if _, ok := expr.Value.(*TupleExpression); !ok {
		t.Fatalf("Expected tuple 'n, m'")
	}
}

func TestCheckAssignmentToTuple(t *testing.T) {
	parser := MakeParser(nil)
	parser.scope.Add("a", Loc{}, Number{})
	parser.scope.Add("b", Loc{}, String{})
	assignment := &Assignment{
		Pattern: &TupleExpression{Elements: []Expression{
			&Identifier{Token: literal{kind: Name, value: "a"}},
			&Identifier{Token: literal{kind: Name, value: "b"}},
		}},
		Value: &TupleExpression{Elements: []Expression{
			&Literal{literal{kind: NumberLiteral, value: "42"}},
			&Literal{literal{kind: StringLiteral, value: "\"Hi!\""}},
		}},
		Operator: token{kind: Assign},
	}
	assignment.typeCheck(parser)
	if len(parser.errors) != 0 {
		t.Fatalf("Expected no errors, got %#v", parser.errors)
	}
}

func TestCheckAssignmentToTupleBadType(t *testing.T) {
	parser := MakeParser(nil)
	parser.scope.Add("a", Loc{}, Number{})
	parser.scope.Add("b", Loc{}, String{})
	assignment := &Assignment{
		Pattern: &TupleExpression{Elements: []Expression{
			&Identifier{Token: literal{kind: Name, value: "a"}},
			&Identifier{Token: literal{kind: Name, value: "b"}},
		}},
		Value: &TupleExpression{Elements: []Expression{
			&Literal{literal{kind: StringLiteral, value: "\"Hi!\""}},
			&Literal{literal{kind: NumberLiteral, value: "42"}},
		}},
		Operator: token{kind: Assign},
	}
	assignment.typeCheck(parser)
	if len(parser.errors) != 1 {
		t.Fatalf("Expected 1 error, got %#v", parser.errors)
	}
}

func TestParseAssignmentToMap(t *testing.T) {
	parser := MakeParser(strings.NewReader("map[\"key\"] = 42"))
	node := parser.parseAssignment()

	if len(parser.errors) > 0 {
		t.Fatalf("Expected no errors, got %#v", parser.errors)
	}
	assignment, ok := node.(*Assignment)
	if !ok {
		t.Fatal("Assignment expected")
	}
	if _, ok := assignment.Pattern.(*ComputedAccessExpression); !ok {
		t.Fatalf("Computed access expected")
	}
}

func TestCheckAssignmentToMap(t *testing.T) {
	parser := MakeParser(nil)
	parser.scope.Add("map", Loc{}, makeMapType(Number{}, Number{}))
	declaration := &Assignment{
		Pattern: &ComputedAccessExpression{
			Expr: &Identifier{Token: literal{kind: Name, value: "map"}},
			Property: &BracketedExpression{
				Expr: &Literal{literal{kind: NumberLiteral, value: "42"}},
			},
		},
		Value:    &Literal{literal{kind: NumberLiteral, value: "42"}},
		Operator: token{kind: Assign},
	}
	declaration.typeCheck(parser)
	if len(parser.errors) != 0 {
		t.Fatalf("Expected no errors, got %#v", parser.errors)
	}
}

func TestCheckAssignmentToMapBadKey(t *testing.T) {
	parser := MakeParser(nil)
	parser.scope.Add("map", Loc{}, makeMapType(Number{}, Number{}))
	declaration := &Assignment{
		Pattern: &ComputedAccessExpression{
			Expr: &Identifier{Token: literal{kind: Name, value: "map"}},
			Property: &BracketedExpression{
				Expr: &Literal{literal{kind: StringLiteral, value: "\"key\""}},
			},
		},
		Value:    &Literal{literal{kind: NumberLiteral, value: "42"}},
		Operator: token{kind: Assign},
	}
	declaration.typeCheck(parser)
	if len(parser.errors) != 1 {
		t.Fatalf("Expected 1 error, got %#v", parser.errors)
	}
}

func TestCheckAssignmentToMapBadValue(t *testing.T) {
	parser := MakeParser(nil)
	parser.scope.Add("map", Loc{}, makeMapType(Number{}, Number{}))
	declaration := &Assignment{
		Pattern: &ComputedAccessExpression{
			Expr: &Identifier{Token: literal{kind: Name, value: "map"}},
			Property: &BracketedExpression{
				Expr: &Literal{literal{kind: NumberLiteral, value: "42"}},
			},
		},
		Value:    &Literal{literal{kind: StringLiteral, value: "\"42\""}},
		Operator: token{kind: Assign},
	}
	declaration.typeCheck(parser)
	if len(parser.errors) != 1 {
		t.Fatalf("Expected 1 error, got %#v", parser.errors)
	}
}

func TestCheckVariableDeclaration(t *testing.T) {
	parser := MakeParser(nil)
	declaration := &Assignment{
		Pattern:  &Identifier{Token: literal{kind: Name, value: "v"}},
		Value:    &Literal{literal{kind: NumberLiteral, value: "42"}},
		Operator: token{kind: Declare},
	}
	declaration.typeCheck(parser)
	if len(parser.errors) != 0 {
		t.Fatalf("Expected no errors, got %#v", parser.errors)
	}

	v, ok := parser.scope.Find("v")
	if !ok {
		t.Fatalf("Expected 'v' to have been declared as a number (got %#v)", v)
	}
	if _, ok := v.Typing.(Number); !ok {
		t.Fatalf("Expected 'v' to have been declared as a number (got %#v)", v)
	}
}

func TestCheckVariableDeclarationNoNil(t *testing.T) {
	parser := MakeParser(nil)
	// v := ()
	declaration := &Assignment{
		Pattern:  &Identifier{Token: literal{kind: Name, value: "v"}},
		Value:    &ParenthesizedExpression{},
		Operator: token{kind: Declare},
	}
	declaration.typeCheck(parser)
	if len(parser.errors) != 1 {
		t.Fatalf("Expected 1 error, got %v: %#v", len(parser.errors), parser.errors)
	}
}

func TestCheckTupleDeclaration(t *testing.T) {
	parser := MakeParser(nil)
	declaration := &Assignment{
		Pattern: &TupleExpression{Elements: []Expression{
			&Identifier{Token: literal{kind: Name, value: "a"}},
			&Identifier{Token: literal{kind: Name, value: "b"}},
		}},
		Value: &TupleExpression{Elements: []Expression{
			&Literal{literal{kind: NumberLiteral, value: "42"}},
			&Literal{literal{kind: StringLiteral, value: "\"Hi!\""}},
		}},
		Operator: token{kind: Declare},
	}
	declaration.typeCheck(parser)
	if len(parser.errors) != 0 {
		t.Fatalf("Expected no errors, got %#v", parser.errors)
	}

	a, ok := parser.scope.Find("a")
	if !ok {
		t.Fatalf("Expected 'a' to have been declared as a number (got %#v)", a)
	}
	if _, ok := a.Typing.(Number); !ok {
		t.Fatalf("Expected 'a' to have been declared as a number (got %#v)", a)
	}

	b, ok := parser.scope.Find("b")
	if !ok {
		t.Fatalf("Expected 'b' to have been declared as a string (got %#v)", b)
	}
	if _, ok := b.Typing.(String); !ok {
		t.Fatalf("Expected 'b' to have been declared as a string (got %#v)", b)
	}
}

func TestCheckTupleDeclarationBadInit(t *testing.T) {
	parser := MakeParser(nil)
	declaration := &Assignment{
		Pattern: &TupleExpression{Elements: []Expression{
			&Identifier{Token: literal{kind: Name, value: "a"}},
			&Identifier{Token: literal{kind: Name, value: "b"}},
		}},
		Value:    &Literal{literal{kind: NumberLiteral, value: "42"}},
		Operator: token{kind: Declare},
	}
	declaration.typeCheck(parser)
	if len(parser.errors) != 1 {
		t.Fatalf("Expected 1 error, got %#v", parser.errors)
	}
}

func TestCheckTupleDeclarationTooMany(t *testing.T) {
	parser := MakeParser(nil)
	declaration := &Assignment{
		Pattern: &TupleExpression{Elements: []Expression{
			&Identifier{Token: literal{kind: Name, value: "a"}},
			&Identifier{Token: literal{kind: Name, value: "b"}},
			&Identifier{Token: literal{kind: Name, value: "c"}},
		}},
		Value: &TupleExpression{Elements: []Expression{
			&Literal{literal{kind: NumberLiteral, value: "42"}},
			&Literal{literal{kind: StringLiteral, value: "\"Hi!\""}},
		}},
		Operator: token{kind: Declare},
	}
	declaration.typeCheck(parser)
	if len(parser.errors) != 1 {
		t.Fatalf("Expected 1 error, got %#v", parser.errors)
	}
}

func TestObjectTypeDefinition(t *testing.T) {
	str := "Type :: {\n"
	str += "    n number\n"
	str += "    s string\n"
	str += "}"
	parser := MakeParser(strings.NewReader(str))
	node := parser.parseAssignment()

	if len(parser.errors) > 0 {
		t.Fatalf("Expected no parsing errors, got:\n%#v", parser.errors)
	}

	expr, ok := node.(*Assignment)
	if !ok {
		t.Fatalf("Expected Assignment, got %#v", node)
	}
	if _, ok := expr.Pattern.(*Identifier); !ok {
		t.Fatalf("Expected identifier 'Type'")
	}
	b, ok := expr.Value.(*Block)
	if !ok {
		t.Fatalf("Expected block, got:\n %#v", expr.Value)
	}
	if _, ok := b.Statements[0].(*Param); !ok {
		t.Fatalf("Expected param, got:\n %#v", b.Statements[0])
	}
}

func TestObjectTypeDefinitionWithDefaults(t *testing.T) {
	str := "Type :: {\n"
	str += "    n number\n"
	str += "    d: 0\n"
	str += "}"
	parser := MakeParser(strings.NewReader(str))
	parser.parseAssignment()

	if len(parser.errors) > 0 {
		t.Fatalf("Expected no parsing errors, got:\n%#v", parser.errors)
	}
}

func TestCheckObjectTypeDefinition(t *testing.T) {
	parser := MakeParser(nil)
	declaration := &Assignment{
		Pattern: &Identifier{Token: literal{kind: Name, value: "Type"}},
		Value: &Block{Statements: []Node{
			&Param{
				Identifier: &Identifier{Token: literal{kind: Name, value: "key"}},
				Complement: &Literal{token{kind: NumberKeyword}},
			},
			&Param{
				Identifier: &Identifier{Token: literal{kind: Name, value: "optional"}},
				Complement: &UnaryExpression{
					Operator: token{kind: QuestionMark},
					Operand:  &Literal{token{kind: NumberKeyword}},
				},
			},
			&Entry{
				Key:   &Identifier{Token: literal{kind: Name, value: "default"}},
				Value: &Literal{literal{kind: NumberLiteral, value: "0"}},
			},
		}},
		Operator: token{kind: Define},
	}
	declaration.typeCheck(parser)

	if len(parser.errors) > 0 {
		t.Fatalf("Expected no errors, got %#v", parser.errors)
	}

	ty, ok := parser.scope.Find("Type")
	if !ok {
		t.Fatal("Expected 'Type' to have been added to scope")
	}
	typing, ok := ty.Typing.(Type)
	if !ok {
		t.Fatal("Expected type 'Type'")
	}
	alias, ok := typing.Value.(TypeAlias)
	if !ok {
		t.Fatal("Expected an alias")
	}
	object, ok := alias.Ref.(Object)
	if !ok {
		t.Fatal("Expected an object")
	}
	_ = object
}

func TestCheckFunctionDefinition(t *testing.T) {
	parser := MakeParser(nil)
	declaration := &Assignment{
		Pattern: &Identifier{Token: literal{kind: Name, value: "function"}},
		Value: &FunctionExpression{
			Params: &ParenthesizedExpression{Expr: &TupleExpression{}},
			Body: &Block{Statements: []Node{
				&Literal{literal{kind: NumberLiteral, value: "42"}},
			}},
		},
		Operator: token{kind: Define},
	}
	declaration.typeCheck(parser)

	if len(parser.errors) > 0 {
		t.Fatalf("Expected no errors, got %#v", parser.errors)
	}

	ty, ok := parser.scope.Find("function")
	if !ok {
		t.Fatal("Expected 'function' to have been added to scope")
	}
	function, ok := ty.Typing.(Function)
	if !ok {
		t.Fatalf("Expected a function, got %v", ty.Typing.Text())
	}
	_ = function
}

func TestCheckGenericTypeDefinition(t *testing.T) {
	parser := MakeParser(nil)
	declaration := &Assignment{
		Pattern: &ComputedAccessExpression{
			Expr: &Identifier{Token: literal{kind: Name, value: "Boxed"}},
			Property: &BracketedExpression{Expr: &TupleExpression{Elements: []Expression{
				&Param{Identifier: &Identifier{Token: literal{kind: Name, value: "Type"}}},
			}}},
		},
		Value: &Block{Statements: []Node{
			&Param{
				Identifier: &Identifier{Token: literal{kind: Name, value: "value"}},
				Complement: &Identifier{Token: literal{kind: Name, value: "Type"}},
			},
		}},
		Operator: token{kind: Define},
	}
	declaration.typeCheck(parser)

	if len(parser.errors) > 0 {
		t.Fatalf("Expected no errors, got %#v", parser.errors)
	}

	ty, ok := parser.scope.Find("Boxed")
	if !ok {
		t.Fatal("Expected 'Boxed' to have been added to scope")
	}
	typing, ok := ty.Typing.(Type)
	if !ok {
		t.Fatal("Expected type 'Type'")
	}
	alias, ok := typing.Value.(TypeAlias)
	if !ok {
		t.Fatal("Expected an alias")
	}
	if len(alias.Params) != 1 {
		t.Fatalf("Expected 1 type param, got %v", len(alias.Params))
	}
}

func TestMethodDeclaration(t *testing.T) {
	str := "(t Type).method :: () => { t }"
	parser := MakeParser(strings.NewReader(str))
	node := parser.parseAssignment()

	expr, ok := node.(*Assignment)
	if !ok {
		t.Fatalf("Expected Assignment, got %#v", node)
	}
	if _, ok := expr.Pattern.(*PropertyAccessExpression); !ok {
		t.Fatalf("Expected method declaration")
	}
	if _, ok := expr.Value.(*FunctionExpression); !ok {
		t.Fatalf("Expected FunctionExpression, got %#v", expr.Value)
	}
}

func TestCheckMethodDeclaration(t *testing.T) {
	parser := MakeParser(nil)
	parser.scope.Add(
		"Type",
		Loc{},
		Type{TypeAlias{Name: "Type", Ref: Number{}}},
	)
	// (t Type).method :: () => { t }
	node := &Assignment{
		Pattern: &PropertyAccessExpression{
			Expr: &ParenthesizedExpression{Expr: &Param{
				Identifier: &Identifier{Token: literal{kind: Name, value: "t"}},
				Complement: &Identifier{Token: literal{kind: Name, value: "Type"}},
			}},
			Property: &Identifier{Token: literal{kind: Name, value: "method"}},
		},
		Value: &FunctionExpression{
			Params: &ParenthesizedExpression{Expr: &TupleExpression{}},
			Body: &Block{Statements: []Node{
				&Identifier{Token: literal{kind: Name, value: "t"}},
			}},
		},
		Operator: token{kind: Define},
	}
	node.typeCheck(parser)

	if len(parser.errors) > 0 {
		t.Fatalf("Expected no errors, got %#v", parser.errors)
	}

	v, _ := parser.scope.Find("Type")
	alias := v.Typing.(Type).Value.(TypeAlias)
	method, ok := alias.Methods["method"]
	if !ok {
		t.Fatal("Expected method to have been declared")
	}
	_ = method
}
