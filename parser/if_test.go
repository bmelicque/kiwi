package parser

import "testing"

func TestIf(t *testing.T) {
	// if n == 2 { return 1 }
	tokenizer := testTokenizer{tokens: []Token{
		token{kind: IfKeyword},
		literal{kind: Name, value: "n"},
		token{kind: Equal},
		literal{kind: NumberLiteral, value: "2"},
		token{kind: LeftBrace},
		token{kind: ReturnKeyword},
		literal{kind: NumberLiteral, value: "1"},
		token{kind: RightBrace},
	}}
	parser := MakeParser(&tokenizer)
	node := parser.parseIfExpression()
	if node.Body == nil {
		t.Fatalf("Expected a body, got %#v", node)
	}
	alias, ok := node.Type().(TypeAlias)
	if !ok || alias.Name != "?" {
		t.Fatalf("Expected option type")
	}
}

func TestIfWithNonBoolean(t *testing.T) {
	// if 42 { }
	parser := MakeParser(nil)
	expr := IfExpression{
		Condition: &Literal{Token: literal{kind: NumberLiteral, value: "42"}},
		Body:      &Block{},
	}
	expr.typeCheck(parser)
	if len(parser.errors) != 1 {
		t.Fatalf("Expected 1 error, got %#v", parser.errors)
	}
}

func TestIfElse(t *testing.T) {
	// if false { true } else { false }
	tokenizer := testTokenizer{tokens: []Token{
		token{kind: IfKeyword},
		literal{kind: BooleanLiteral, value: "false"},
		token{kind: LeftBrace},
		literal{kind: BooleanLiteral, value: "true"},
		token{kind: RightBrace},
		token{kind: ElseKeyword},
		token{kind: LeftBrace},
		literal{kind: BooleanLiteral, value: "false"},
		token{kind: RightBrace},
	}}
	parser := MakeParser(&tokenizer)
	node := parser.parseIfExpression()

	if len(parser.errors) != 0 {
		t.Fatalf("Expected no errors, got %#v", parser.errors)
	}

	if node.Body == nil {
		t.Fatal("Expected a body")
	}
	if node.Alternate == nil {
		t.Fatal("Expected alternate")
	}
	if _, ok := node.Alternate.(*Block); !ok {
		t.Fatalf("Expected body alternate, got %#v", node.Alternate)
	}
	if _, ok := node.Type().(Boolean); !ok {
		t.Fatalf("Expected a boolean")
	}
}

func TestIfElseWithTypeMismatch(t *testing.T) {
	// if false { 42 } else { false }
	parser := MakeParser(nil)
	expr := IfExpression{
		Keyword:   token{kind: IfKeyword},
		Condition: &Literal{literal{kind: BooleanLiteral, value: "false"}},
		Body: &Block{Statements: []Node{
			&Literal{literal{kind: NumberLiteral, value: "42"}},
		}},
		Alternate: &Block{Statements: []Node{
			&Literal{literal{kind: BooleanLiteral, value: "false"}},
		}},
	}
	expr.typeCheck(parser)

	if len(parser.errors) != 1 {
		t.Fatalf("Expected 1 error, got %#v", parser.errors)
	}
}

func TestIfElseIf(t *testing.T) {
	// if false {} else if true {}
	tokenizer := testTokenizer{tokens: []Token{
		token{kind: IfKeyword},
		literal{kind: BooleanLiteral, value: "false"},
		token{kind: LeftBrace},
		token{kind: RightBrace},
		token{kind: ElseKeyword},
		token{kind: IfKeyword},
		literal{kind: BooleanLiteral, value: "true"},
		token{kind: LeftBrace},
		token{kind: RightBrace},
	}}
	parser := MakeParser(&tokenizer)
	node := parser.parseIfExpression()

	if len(parser.errors) != 0 {
		t.Fatalf("Expected no errors, got %#v", parser.errors)
	}

	if node.Body == nil {
		t.Fatal("Expected a body")
	}
	if node.Alternate == nil {
		t.Fatal("Expected alternate")
	}
	if _, ok := node.Alternate.(*IfExpression); !ok {
		t.Fatalf("Expected another 'if' as alternate, got %#v", node.Alternate)
	}
}

func TestIfPattern(t *testing.T) {
	// if Some(s) := option {}
	parser := MakeParser(&testTokenizer{tokens: []Token{
		token{kind: IfKeyword},
		literal{kind: Name, value: "Some"},
		token{kind: LeftParenthesis},
		literal{kind: Name, value: "s"},
		token{kind: RightParenthesis},
		token{kind: Declare},
		literal{kind: Name, value: "option"},
		token{kind: LeftBrace},
		token{kind: RightBrace},
	}})
	expr := parser.parseIfExpression()

	if len(parser.errors) != 0 {
		t.Fatalf("Expected no errors, got %#v", parser.errors)
	}

	if _, ok := expr.Condition.(*Assignment); !ok {
		t.Fatalf("Expected assignment, got %#v", expr.Condition)
	}
}

func TestCheckIfPattern(t *testing.T) {
	parser := MakeParser(nil)
	parser.scope.Add("option", Loc{}, makeOptionType(Number{}))
	// if Some(s) := option { s } else { 0 }
	expr := &IfExpression{
		Condition: &Assignment{
			Pattern: &CallExpression{
				Callee: &Identifier{Token: literal{kind: Name, value: "Some"}},
				Args: &ParenthesizedExpression{
					Expr: &Identifier{Token: literal{kind: Name, value: "s"}},
				},
			},
			Value:    &Identifier{Token: literal{kind: Name, value: "option"}},
			Operator: token{kind: Declare},
		},
		Body: &Block{Statements: []Node{
			&Identifier{Token: literal{kind: Name, value: "s"}},
		}},
		Alternate: &Block{Statements: []Node{
			&Literal{literal{kind: NumberLiteral, value: "0"}},
		}},
	}
	expr.typeCheck(parser)

	if len(parser.errors) != 0 {
		t.Fatalf("Expected no errors, got %#v", parser.errors)
	}

	if _, ok := expr.Type().(Number); !ok {
		t.Fatalf("Expected a number, got %#v", expr.Type())
	}
}
