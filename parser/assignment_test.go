package parser

import "testing"

func TestAssignment(t *testing.T) {
	tokenizer := testTokenizer{tokens: []Token{
		literal{kind: Name, value: "n"},
		token{kind: Assign},
		literal{kind: NumberLiteral, value: "42"},
	}}
	parser := MakeParser(&tokenizer)
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
	parser.scope.Add("value", Loc{}, Primitive{NUMBER})
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
	parser.scope.Add("value", Loc{}, Primitive{NUMBER})
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
	tokenizer := testTokenizer{tokens: []Token{
		literal{kind: Name, value: "n"},
		token{kind: Comma},
		literal{kind: Name, value: "m"},
		token{kind: Assign},
		literal{kind: NumberLiteral, value: "1"},
		token{kind: Comma},
		literal{kind: NumberLiteral, value: "2"},
	}}
	parser := MakeParser(&tokenizer)
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
	parser.scope.Add("a", Loc{}, Primitive{NUMBER})
	parser.scope.Add("b", Loc{}, Primitive{STRING})
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
	parser.scope.Add("a", Loc{}, Primitive{NUMBER})
	parser.scope.Add("b", Loc{}, Primitive{STRING})
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
	if !ok || v.typing.Kind() != NUMBER {
		t.Fatalf("Expected 'v' to have been declared as a number (got %#v)", v)
	}
}

func TestObjectDeclaration(t *testing.T) {
	tokenizer := testTokenizer{tokens: []Token{
		literal{kind: Name, value: "Type"},
		token{kind: Define},
		token{kind: LeftParenthesis},
		token{kind: EOL},

		literal{kind: Name, value: "n"},
		token{kind: NumberKeyword},
		token{kind: Comma},
		token{kind: EOL},

		token{kind: RightParenthesis},
	}}
	parser := MakeParser(&tokenizer)
	node := parser.parseAssignment()

	expr, ok := node.(*Assignment)
	if !ok {
		t.Fatalf("Expected Assignment, got %#v", node)
	}
	if _, ok := expr.Pattern.(*Identifier); !ok {
		t.Fatalf("Expected identifier 'Type'")
	}
	if _, ok := expr.Value.(*ParenthesizedExpression); !ok {
		t.Fatalf("Expected ParenthesizedExpression")
	}
}

func TestMethodDeclaration(t *testing.T) {
	tokenizer := testTokenizer{tokens: []Token{
		token{kind: LeftParenthesis},
		literal{kind: Name, value: "t"},
		literal{kind: Name, value: "Type"},
		token{kind: RightParenthesis},
		token{kind: Dot},
		literal{kind: Name, value: "method"},
		token{kind: Define},
		token{kind: LeftParenthesis},
		token{kind: RightParenthesis},
		token{kind: SlimArrow},
		token{kind: LeftParenthesis},
		token{kind: RightParenthesis},
	}}
	parser := MakeParser(&tokenizer)
	node := parser.parseAssignment()

	expr, ok := node.(*Assignment)
	if !ok {
		t.Fatalf("Expected Assignment, got %#v", node)
	}
	if _, ok := expr.Pattern.(*PropertyAccessExpression); !ok {
		t.Fatalf("Expected method declaration")
	}
	if _, ok := expr.Value.(*FunctionTypeExpression); !ok {
		t.Fatalf("Expected FunctionTypeExpression, got %#v", expr.Value)
	}
}
