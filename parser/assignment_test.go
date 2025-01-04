package parser

import (
	"strings"
	"testing"
)

func TestParseAssignment(t *testing.T) {
	parser := MakeParser(strings.NewReader("n = 42"))
	node := parser.parseAssignment()
	loc := Loc{Position{1, 1}, Position{1, 7}}
	if node.Loc() != loc {
		t.Fatalf("Expected loc %v, got %v", node.Loc(), loc)
	}
	testParserErrors(t, parser, 0)

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

func TestParseAssignmentToField(t *testing.T) {
	parser := MakeParser(strings.NewReader("object.key = 42"))
	parser.parseAssignment()
	testParserErrors(t, parser, 0)
}

func TestParseAssignmentBadAssignee(t *testing.T) {
	var parser *Parser

	parser = MakeParser(strings.NewReader("f() = 42"))
	parser.parseAssignment()
	testParserErrors(t, parser, 1)

	parser = MakeParser(strings.NewReader("T{} = 42"))
	parser.parseAssignment()
	testParserErrors(t, parser, 1)

	parser = MakeParser(strings.NewReader("a + b = 42"))
	parser.parseAssignment()
	testParserErrors(t, parser, 1)
}

func TestParseAssignmentShorthand(t *testing.T) {
	parser := MakeParser(strings.NewReader("n += 42"))
	node := parser.parseAssignment()

	testParserErrors(t, parser, 0)

	expr, ok := node.(*Assignment)
	if !ok {
		t.Fatalf("Expected Assignment, got %#v", node)
	}
	if expr.Operator.Kind() != AddAssign {
		t.Fatal("Expected +=, got", expr.Operator)
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

func TestCheckArithmeticAssigns(t *testing.T) {
	parser := MakeParser(nil)
	parser.scope.Add("value", Loc{}, Number{})
	assignment := &Assignment{
		Pattern:  &Identifier{Token: literal{kind: Name, value: "value"}},
		Value:    &Literal{literal{kind: NumberLiteral, value: "42"}},
		Operator: token{kind: AddAssign},
	}
	assignment.typeCheck(parser)

	assignment.Operator = token{kind: SubAssign}
	assignment.typeCheck(parser)

	assignment.Operator = token{kind: MulAssign}
	assignment.typeCheck(parser)

	assignment.Operator = token{kind: DivAssign}
	assignment.typeCheck(parser)

	assignment.Operator = token{kind: ModAssign}
	assignment.typeCheck(parser)

	testParserErrors(t, parser, 0)
}

func TestCheckArithmeticBadAssign(t *testing.T) {
	parser := MakeParser(nil)
	parser.scope.Add("value", Loc{}, String{})
	assignment := &Assignment{
		Pattern:  &Identifier{Token: literal{kind: Name, value: "value"}},
		Value:    &Literal{literal{kind: NumberLiteral, value: "42"}},
		Operator: token{kind: AddAssign},
	}
	assignment.typeCheck(parser)

	testParserErrors(t, parser, 1)
}

func TestCheckConcatAssign(t *testing.T) {
	parser := MakeParser(nil)
	parser.scope.Add("value", Loc{}, String{})
	assignment := &Assignment{
		Pattern:  &Identifier{Token: literal{kind: Name, value: "value"}},
		Value:    &Literal{literal{kind: StringLiteral, value: "\"\""}},
		Operator: token{kind: ConcatAssign},
	}
	assignment.typeCheck(parser)

	testParserErrors(t, parser, 0)
}

func TestCheckConcatBadAssign(t *testing.T) {
	parser := MakeParser(nil)
	parser.scope.Add("value", Loc{}, Number{})
	assignment := &Assignment{
		Pattern:  &Identifier{Token: literal{kind: Name, value: "value"}},
		Value:    &Literal{literal{kind: StringLiteral, value: "\"\""}},
		Operator: token{kind: ConcatAssign},
	}
	assignment.typeCheck(parser)

	testParserErrors(t, parser, 1)
}

func TestCheckLogicalAssign(t *testing.T) {
	parser := MakeParser(nil)
	parser.scope.Add("value", Loc{}, Boolean{})
	assignment := &Assignment{
		Pattern:  &Identifier{Token: literal{kind: Name, value: "value"}},
		Value:    &Literal{literal{kind: BooleanLiteral, value: "true"}},
		Operator: token{kind: LogicalAndAssign},
	}
	assignment.typeCheck(parser)

	assignment.Operator = token{kind: LogicalOrAssign}
	assignment.typeCheck(parser)

	testParserErrors(t, parser, 0)
}

func TestCheckLogicalBadAssign(t *testing.T) {
	parser := MakeParser(nil)
	parser.scope.Add("value", Loc{}, Number{})
	assignment := &Assignment{
		Pattern:  &Identifier{Token: literal{kind: Name, value: "value"}},
		Value:    &Literal{literal{kind: BooleanLiteral, value: "true"}},
		Operator: token{kind: LogicalAndAssign},
	}
	assignment.typeCheck(parser)

	testParserErrors(t, parser, 1)
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
	b, ok := expr.Value.(*BracedExpression)
	if !ok {
		t.Fatalf("Expected *BracedExpression, got:\n %#v", expr.Value)
	}
	elements := b.Expr.(*TupleExpression).Elements
	if _, ok := elements[0].(*Param); !ok {
		t.Fatalf("Expected param, got:\n %#v", elements[0])
	}
}

func TestObjectTypeDefinitionSingleLine(t *testing.T) {
	str := "Type :: { n number, s string }"
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
	b, ok := expr.Value.(*BracedExpression)
	if !ok {
		t.Fatalf("Expected block, got:\n %#v", expr.Value)
	}
	elements := b.Expr.(*TupleExpression).Elements
	if _, ok := elements[0].(*Param); !ok {
		t.Fatalf("Expected param, got:\n %#v", elements[0])
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

func TestObjectTypeDefinitionWithDuplicates(t *testing.T) {
	str := "Type :: {\n"
	str += "    n number\n"
	str += "    n string\n"
	str += "}"
	parser := MakeParser(strings.NewReader(str))
	parser.parseAssignment()

	if len(parser.errors) != 2 {
		t.Fatalf("Expected 2 parsing errors, got:\n%#v", parser.errors)
	}
}

func TestParseGenericObjectDefinition(t *testing.T) {
	source := "Boxed[Type] :: {\n"
	source += "    value Type\n"
	source += "}"
	parser := MakeParser(strings.NewReader(source))
	a := parser.parseAssignment()
	a.typeCheck(parser)
	testParserErrors(t, parser, 0)

	v, ok := parser.scope.Find("Boxed")
	if !ok {
		t.Fatalf("Expected to find 'Boxed' in scope")
	}
	alias, ok := v.Typing.(Type).Value.(TypeAlias)
	if !ok || alias.Name != "Boxed" {
		t.Fatalf("Expected 'Boxed' type, got %v", v.Typing.Text())
	}
	if len(alias.Params) != 1 {
		t.Fatal("Expected 1 param")
	}
	member := alias.Ref.(Object).Members[0]
	if _, ok := member.Type.(Generic); ok {
		t.Fatalf("Expected generic type")
	}
}

func TestCheckObjectTypeDefinition(t *testing.T) {
	parser := MakeParser(nil)
	declaration := &Assignment{
		Pattern: &Identifier{Token: literal{kind: Name, value: "Type"}},
		Value: &BracedExpression{Expr: &TupleExpression{Elements: []Expression{
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
		}}},
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

func TestCheckObjectTypeDefinitionBadOrder(t *testing.T) {
	parser := MakeParser(nil)
	declaration := &Assignment{
		Pattern: &Identifier{Token: literal{kind: Name, value: "Type"}},
		Value: &BracedExpression{Expr: &TupleExpression{Elements: []Expression{
			&Entry{
				Key:   &Identifier{Token: literal{kind: Name, value: "default"}},
				Value: &Literal{literal{kind: NumberLiteral, value: "0"}},
			},
			&Param{
				Identifier: &Identifier{Token: literal{kind: Name, value: "key"}},
				Complement: &Literal{token{kind: NumberKeyword}},
			},
		}}},
		Operator: token{kind: Define},
	}
	declaration.typeCheck(parser)

	if len(parser.errors) != 1 {
		t.Fatalf("Expected 1 error, got %#v", parser.errors)
	}
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
		Value: &BracedExpression{Expr: &TupleExpression{Elements: []Expression{
			&Param{
				Identifier: &Identifier{Token: literal{kind: Name, value: "value"}},
				Complement: &Identifier{Token: literal{kind: Name, value: "Type"}},
			},
		}}},
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

// -----------------------
// TEST METHOD DEFINITIONS
// -----------------------

func TestParseMethodDeclaration(t *testing.T) {
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

	testParserErrors(t, parser, 0)

	v, _ := parser.scope.Find("Type")
	alias := v.Typing.(Type).Value.(TypeAlias)
	method, ok := alias.Methods["method"]
	if !ok {
		t.Fatal("Expected method to have been declared")
	}
	params := method.(Function).Params.Elements
	if len(params) != 0 {
		t.Fatalf("Expected no params for method, found %#v", params)
	}
}
