package parser

import "testing"

func TestParseCatchExpression(t *testing.T) {
	parser := MakeParser(&testTokenizer{tokens: []Token{
		literal{kind: Name, value: "result"},
		token{kind: CatchKeyword},
		literal{kind: Name, value: "err"},
		token{kind: LeftBrace},
		literal{kind: BooleanLiteral, value: "true"},
		token{kind: RightBrace},
	}})
	expr := parser.parseCatchExpression()

	if len(parser.errors) > 0 {
		t.Fatalf("Expected no errors, got %#v", parser.errors)
	}
	if _, ok := expr.(*CatchExpression); !ok {
		t.Fatalf("Expected 'catch' expression, got %#v", expr)
	}
}

func TestParseCatchExpressionNoIdentifier(t *testing.T) {
	parser := MakeParser(&testTokenizer{tokens: []Token{
		literal{kind: Name, value: "result"},
		token{kind: CatchKeyword},
		token{kind: LeftBrace},
		token{kind: RightBrace},
	}})
	parser.parseCatchExpression()

	if len(parser.errors) != 1 {
		t.Fatalf("Expected 1 error, got %v: %#v", len(parser.errors), parser.errors)
	}
}

func TestParseCatchExpressionBadIdentifier(t *testing.T) {
	parser := MakeParser(&testTokenizer{tokens: []Token{
		literal{kind: Name, value: "result"},
		token{kind: CatchKeyword},
		token{kind: NumberKeyword},
		token{kind: LeftBrace},
		token{kind: RightBrace},
	}})
	parser.parseCatchExpression()

	if len(parser.errors) != 1 {
		t.Fatalf("Expected 1 error, got %v: %#v", len(parser.errors), parser.errors)
	}
}

func TestParseCatchExpressionBadTokens(t *testing.T) {
	parser := MakeParser(&testTokenizer{tokens: []Token{
		literal{kind: Name, value: "result"},
		token{kind: CatchKeyword},
		literal{kind: Name, value: "err"},
		literal{kind: Name, value: "err"},
		token{kind: LeftBrace},
		token{kind: RightBrace},
	}})
	parser.parseCatchExpression()

	if len(parser.errors) != 1 {
		t.Fatalf("Expected 1 error, got %v: %#v", len(parser.errors), parser.errors)
	}
}

func TestCheckCatchExpression(t *testing.T) {
	parser := MakeParser(nil)
	parser.scope.Add(
		"result",
		Loc{},
		makeResultType(Number{}, String{}),
	)
	expr := &CatchExpression{
		Left:       &Identifier{Token: literal{kind: Name, value: "result"}},
		Keyword:    token{kind: CatchKeyword},
		Identifier: &Identifier{Token: literal{kind: Name, value: "err"}},
		Body: &Block{Statements: []Node{
			&Identifier{Token: literal{kind: Name, value: "err"}},
			&Literal{literal{kind: NumberLiteral, value: "0"}},
		}},
	}
	expr.typeCheck(parser)

	if len(parser.errors) > 0 {
		t.Fatalf("Expected no errors, got %#v", parser.errors)
	}
	if _, ok := expr.Type().(Number); !ok {
		t.Fatalf("Expected number")
	}
}

func TestCheckCatchExpressionNotResult(t *testing.T) {
	parser := MakeParser(nil)
	expr := &CatchExpression{
		Left:       &Literal{literal{kind: NumberLiteral, value: "42"}},
		Keyword:    token{kind: CatchKeyword},
		Identifier: &Identifier{Token: literal{kind: Name, value: "err"}},
		Body: &Block{Statements: []Node{
			&Identifier{Token: literal{kind: Name, value: "err"}},
			&Literal{literal{kind: NumberLiteral, value: "0"}},
		}},
	}
	expr.typeCheck(parser)

	if len(parser.errors) != 1 {
		t.Fatalf("Expected 1 error, got %v: %#v", len(parser.errors), parser.errors)
	}
	if _, ok := expr.Type().(Number); !ok {
		t.Fatalf("Expected number")
	}
}

func TestCheckCatchExpressionBlockNotMatching(t *testing.T) {
	parser := MakeParser(nil)
	parser.scope.Add(
		"result",
		Loc{},
		makeResultType(Number{}, String{}),
	)
	expr := &CatchExpression{
		Left:       &Identifier{Token: literal{kind: Name, value: "result"}},
		Keyword:    token{kind: CatchKeyword},
		Identifier: &Identifier{Token: literal{kind: Name, value: "err"}},
		Body: &Block{Statements: []Node{
			&Identifier{Token: literal{kind: Name, value: "err"}},
		}},
	}
	expr.typeCheck(parser)

	if len(parser.errors) != 1 {
		t.Fatalf("Expected 1 error, got %v: %#v", len(parser.errors), parser.errors)
	}
	if _, ok := expr.Body.Type().(String); !ok {
		t.Fatalf("Expected string, got %#v", expr.Type())
	}
}
