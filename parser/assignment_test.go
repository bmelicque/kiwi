package parser

import (
	"reflect"
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

func TestCheckAssignmentPattern(t *testing.T) {
	type test struct {
		name      string
		pattern   Expression
		wantError bool
	}
	tests := []test{
		{
			name: "is module access",
			pattern: &PropertyAccessExpression{
				Expr:     &Identifier{Token: literal{kind: Name, value: "module"}},
				Property: &Identifier{Token: literal{kind: Name, value: "member"}},
			},
			wantError: true,
		},
		{
			name: "is access to non-module",
			pattern: &PropertyAccessExpression{
				Expr:     &Identifier{Token: literal{kind: Name, value: "object"}},
				Property: &Identifier{Token: literal{kind: Name, value: "member"}},
			},
			wantError: false,
		},
		{
			name:      "is not access",
			pattern:   &Identifier{Token: literal{kind: Name, value: "object"}},
			wantError: false,
		},
	}

	scope := NewScope(ProgramScope)
	scope.Add("module", Loc{}, Module{Object{Members: []ObjectMember{
		{"member", Number{}},
	}}})
	scope.Add("object", Loc{}, TypeAlias{
		Name: "Object",
		Ref: Object{Members: []ObjectMember{
			{"member", Number{}},
		}},
		Methods: map[string]ExpressionType{},
	})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := MakeParser(strings.NewReader(""))
			p.scope = scope

			a := &Assignment{Pattern: tt.pattern, Operator: token{kind: Assign}}
			checkAssignmentPattern(p, a)

			if tt.wantError && len(p.errors) == 0 {
				t.Error("expected error, got none")
			}
			if !tt.wantError && len(p.errors) > 0 {
				t.Errorf("expected no error, got one: %v", p.errors[0].Text())
			}
		})
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
	// v := {}
	declaration := &Assignment{
		Pattern:  &Identifier{Token: literal{kind: Name, value: "v"}},
		Value:    &Block{},
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

func testObjectTypeDefinition(t *testing.T, str string) {
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

	expr.typeCheck(parser)
	testParserErrors(t, parser, 0)
	v, _ := parser.scope.Find("Type")
	if v == nil {
		t.Fatal("expected to find variable 'Type'")
	}
	if _, ok := v.Typing.(Type); !ok {
		t.Fatal("expected 'Type' to be a type")
	}
}

func TestObjectTypeDefinitionMultiline(t *testing.T) {
	str := "Type :: {\n"
	str += "    n number\n"
	str += "    s string\n"
	str += "}"
	testObjectTypeDefinition(t, str)
}

func TestObjectTypeDefinitionSingleLine(t *testing.T) {
	str := "Type :: { n number, s string }"
	testObjectTypeDefinition(t, str)
}

func TestObjectTypeDefinitionWithDefaults(t *testing.T) {
	str := "Type :: {\n"
	str += "    n number\n"
	str += "    d: 0\n"
	str += "}"
	parser := MakeParser(strings.NewReader(str))
	parser.parseAssignment().typeCheck(parser)
	testParserErrors(t, parser, 0)
}

func TestObjectTypeDefinitionWithDuplicates(t *testing.T) {
	str := "Type :: {\n"
	str += "    n number\n"
	str += "    n string\n"
	str += "}"
	parser := MakeParser(strings.NewReader(str))
	parser.parseAssignment().typeCheck(parser)
	testParserErrors(t, parser, 2)
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
	access, ok := expr.Pattern.(*PropertyAccessExpression)
	if !ok {
		t.Fatalf("Expected method declaration")
	}
	if _, ok := access.Expr.(*ParenthesizedExpression); !ok {
		t.Fatalf("Expected parenthesized, got %v", reflect.TypeOf(access.Expr))
	}
	if _, ok := expr.Value.(*FunctionExpression); !ok {
		t.Fatalf("Expected FunctionExpression, got %#v", expr.Value)
	}
}

func TestGetValidatedMethodReceiver(t *testing.T) {
	tests := []struct {
		name          string
		receiver      Expression
		expectedError ErrorKind
		shouldBeNil   bool
	}{
		{
			name: "Valid receiver",
			receiver: &ParenthesizedExpression{
				Expr: &Param{
					Complement: &Identifier{Token: literal{kind: Name, value: "Type"}},
				},
			},
			expectedError: NoError,
			shouldBeNil:   false,
		},
		{
			name:          "Nil receiver",
			receiver:      nil,
			expectedError: ReceiverExpected,
			shouldBeNil:   true,
		},
		{
			name:          "Non-parenthesized expression",
			receiver:      &Identifier{},
			expectedError: ReceiverExpected,
			shouldBeNil:   true,
		},
		{
			name: "Parenthesized but not param",
			receiver: &ParenthesizedExpression{
				Expr: &Identifier{},
			},
			expectedError: ReceiverExpected,
			shouldBeNil:   true,
		},
		{
			name: "Invalid type identifier",
			receiver: &ParenthesizedExpression{
				Expr: &Param{
					Complement: &Identifier{Token: literal{kind: Name, value: "name"}},
				},
			},
			expectedError: TypeIdentifierExpected,
			shouldBeNil:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := MakeParser(strings.NewReader(""))

			result := getValidatedMethodReceiver(p, tt.receiver)

			if tt.expectedError != NoError {
				if len(p.errors) == 0 {
					t.Errorf("expected error %v, got none", tt.expectedError)
				} else if p.errors[0].Kind != tt.expectedError {
					t.Errorf("expected error %v, got %v", tt.expectedError, p.errors[0].Kind)
				}
			}

			if tt.shouldBeNil && result != nil {
				t.Errorf("expected nil result, got %v", result)
			}
		})
	}
}

func TestGetValidatedMethodIdentifier(t *testing.T) {
	tests := []struct {
		name          string
		method        Expression
		expectedError ErrorKind
		shouldBeNil   bool
	}{
		{
			name:          "Valid identifier",
			method:        &Identifier{Token: literal{kind: Name, value: "method"}},
			expectedError: NoError,
			shouldBeNil:   false,
		},
		{
			name:          "Nil method",
			method:        nil,
			expectedError: IdentifierExpected,
			shouldBeNil:   true,
		},
		{
			name:          "Non-identifier expression",
			method:        &ParenthesizedExpression{},
			expectedError: IdentifierExpected,
			shouldBeNil:   true,
		},
		{
			name:          "Type identifier instead of value",
			method:        &Identifier{Token: literal{kind: Name, value: "Method"}},
			expectedError: ValueIdentifierExpected,
			shouldBeNil:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := MakeParser(strings.NewReader(""))

			result := getValidatedMethodIdentifier(p, tt.method)

			if tt.expectedError != NoError {
				if len(p.errors) == 0 {
					t.Errorf("expected error %v, got none", tt.expectedError)
				} else if p.errors[0].Kind != tt.expectedError {
					t.Errorf("expected error %v, got %v", tt.expectedError, p.errors[0].Kind)
				}
			}

			if tt.shouldBeNil && result != nil {
				t.Errorf("expected nil result, got %v", result)
			} else if !tt.shouldBeNil && result == nil {
				t.Error("expected non-nil result, got nil")
			}
		})
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
