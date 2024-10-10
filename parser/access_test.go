package parser

import "testing"

func TestComputedPropertyAccess(t *testing.T) {
	tokenizer := testTokenizer{tokens: []Token{
		literal{kind: Name, value: "n"},
		token{kind: LeftBracket},
		literal{kind: Name, value: "p"},
		token{kind: RightBracket},
	}}
	parser := MakeParser(&tokenizer)
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
	tokenizer := testTokenizer{tokens: []Token{
		literal{kind: Name, value: "n"},
		token{kind: Dot},
		literal{kind: Name, value: "p"},
	}}
	parser := MakeParser(&tokenizer)
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
	tokenizer := testTokenizer{tokens: []Token{
		literal{kind: Name, value: "tuple"},
		token{kind: Dot},
		literal{kind: NumberLiteral, value: "0"},
	}}
	parser := MakeParser(&tokenizer)
	parser.scope.Add("tuple", Loc{}, Tuple{[]ExpressionType{Primitive{NUMBER}}})
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
	tokenizer := testTokenizer{tokens: []Token{
		token{kind: LeftParenthesis},
		literal{kind: Name, value: "t"},
		literal{kind: Name, value: "Type"},
		token{kind: RightParenthesis},
		token{kind: Dot},
		literal{kind: Name, value: "method"},
	}}
	parser := MakeParser(&tokenizer)
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
	// (Self).(method() -> Self)
	tokenizer := testTokenizer{tokens: []Token{
		token{kind: LeftParenthesis},
		literal{kind: Name, value: "Self"},
		token{kind: RightParenthesis},
		token{kind: Dot},
		token{kind: LeftParenthesis},
		literal{kind: Name, value: "method"},
		token{kind: LeftParenthesis},
		token{kind: RightParenthesis},
		token{kind: SlimArrow},
		literal{kind: Name, value: "Self"},
		token{kind: RightParenthesis},
	}}

	parser := MakeParser(&tokenizer)
	node := parser.parseAccessExpression()

	if len(parser.errors) > 0 {
		t.Fatalf("Got %v parsing errors: %#v", len(parser.errors), parser.errors)
	}

	expr, ok := node.(*PropertyAccessExpression)
	if !ok {
		t.Fatalf("Expected PropertyAccessExpression, got %#v", node)
	}

	if _, ok := expr.Expr.(*ParenthesizedExpression); !ok {
		t.Fatalf("Expected ParenthesizedExpression, got %#v", expr.Expr)
	}

	if _, ok := expr.Property.(*ParenthesizedExpression); !ok {
		t.Fatalf("Expected ParenthesizedExpression, got %#v", expr.Property)
	}
}

func TestFunctionCall(t *testing.T) {
	tokenizer := testTokenizer{tokens: []Token{
		literal{kind: Name, value: "f"},
		token{kind: LeftParenthesis},
		literal{kind: NumberLiteral, value: "42"},
		token{kind: RightParenthesis},
	}}
	parser := MakeParser(&tokenizer)
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
	tokenizer := testTokenizer{tokens: []Token{
		literal{kind: Name, value: "f"},
		token{kind: LeftBracket},
		token{kind: NumberKeyword},
		token{kind: RightBracket},
		token{kind: LeftParenthesis},
		literal{kind: NumberLiteral, value: "42"},
		token{kind: RightParenthesis},
	}}
	parser := MakeParser(&tokenizer)
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
	tokenizer := testTokenizer{tokens: []Token{
		literal{kind: Name, value: "Type"},
		token{kind: LeftParenthesis},
		literal{kind: Name, value: "value"},
		token{kind: Colon},
		literal{kind: NumberLiteral, value: "42"},
		token{kind: RightParenthesis},
	}}
	parser := MakeParser(&tokenizer)
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
	tokenizer := testTokenizer{tokens: []Token{
		token{kind: LeftBracket},
		token{kind: RightBracket},
		token{kind: NumberKeyword},
		token{kind: LeftParenthesis},
		literal{kind: NumberLiteral, value: "1"},
		token{kind: Comma},
		literal{kind: NumberLiteral, value: "2"},
		token{kind: RightParenthesis},
	}}
	parser := MakeParser(&tokenizer)
	node := ParseExpression(parser)

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
