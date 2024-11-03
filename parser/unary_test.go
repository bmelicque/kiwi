package parser

import "testing"

func TestUnaryExpression(t *testing.T) {
	// ?number
	parser := MakeParser(&testTokenizer{tokens: []Token{
		token{kind: QuestionMark},
		token{kind: NumberKeyword},
	}})
	expr := parser.parseUnaryExpression()

	if len(parser.errors) != 0 {
		t.Fatalf("Expected no errors, got %#v", parser.errors)
	}
	unary, ok := expr.(*UnaryExpression)
	if !ok {
		t.Fatal("Expected unary expression")
	}
	if unary.Operator.Kind() != QuestionMark {
		t.Fatal("Expected question mark")
	}
	if _, ok := unary.Operand.(*Literal); !ok {
		t.Fatal("Expected literal")
	}
	ty, ok := unary.Type().(Type)
	if !ok {
		t.Fatal("Expected type")
	}
	alias, ok := ty.Value.(TypeAlias)
	if !ok || alias.Name != "?" {
		t.Fatal("Expected option type")
	}
}

func TestCheckOptionType(t *testing.T) {
	// ?number
	parser := MakeParser(nil)
	expr := &UnaryExpression{
		Operator: token{kind: QuestionMark},
		Operand:  &Literal{Token: token{kind: NumberKeyword}},
	}
	expr.typeCheck(parser)

	if len(parser.errors) != 0 {
		t.Fatalf("Expected no errors, got %#v", parser.errors)
	}
	ty, ok := expr.Type().(Type)
	if !ok {
		t.Fatal("Expected type")
	}
	alias, ok := ty.Value.(TypeAlias)
	if !ok || alias.Name != "?" {
		t.Fatal("Expected option type")
	}
	if _, ok := getSomeType(alias.Ref.(Sum)).(Number); !ok {
		t.Fatal("Expected number option type")
	}
}

func TestNoOptionValue(t *testing.T) {
	parser := MakeParser(&testTokenizer{tokens: []Token{
		token{kind: QuestionMark},
		literal{kind: NumberLiteral, value: "42"},
	}})
	expr := parser.parseUnaryExpression()
	expr.typeCheck(parser)

	if len(parser.errors) != 1 {
		t.Fatalf("Expected 1 error, got %#v", parser.errors)
	}
}

func TestCheckErrorType(t *testing.T) {
	// !number
	parser := MakeParser(nil)
	expr := &UnaryExpression{
		Operator: token{kind: Bang},
		Operand:  &Literal{Token: token{kind: NumberKeyword}},
	}
	expr.typeCheck(parser)

	if len(parser.errors) != 0 {
		t.Fatalf("Expected no errors, got %#v", parser.errors)
	}
	ty, ok := expr.Type().(Type)
	if !ok {
		t.Fatal("Expected type")
	}
	alias, ok := ty.Value.(TypeAlias)
	if !ok || alias.Name != "!" {
		t.Fatal("Expected result type")
	}
	if _, ok := alias.Ref.(Sum).getMember("Ok").(Number); !ok {
		t.Fatal("Expected number option type")
	}
}

func TestCheckLogicalNot(t *testing.T) {
	// !true
	parser := MakeParser(nil)
	expr := &UnaryExpression{
		Operator: token{kind: Bang},
		Operand:  &Literal{literal{kind: BooleanLiteral, value: "true"}},
	}
	expr.typeCheck(parser)

	if len(parser.errors) != 0 {
		t.Fatalf("Expected no errors, got %#v", parser.errors)
	}
	if _, ok := expr.Type().(Boolean); !ok {
		t.Fatalf("Expected boolean")
	}

	// !42
	expr.Operand = &Literal{literal{kind: NumberLiteral, value: "42"}}
	expr.typeCheck(parser)
	if len(parser.errors) != 1 {
		t.Fatalf("Expected 1 error, got %#v", parser.errors)
	}
}

func TestListTypeExpression(t *testing.T) {
	tokenizer := testTokenizer{tokens: []Token{
		token{kind: LeftBracket},
		token{kind: RightBracket},
		token{kind: NumberKeyword},
	}}
	parser := MakeParser(&tokenizer)
	node := ParseExpression(parser)

	if len(parser.errors) != 0 {
		t.Fatalf("Expected no errors, got %+v: %#v", len(parser.errors), parser.errors)
	}

	list, ok := node.(*ListTypeExpression)
	if !ok {
		t.Fatalf("Expected ListExpression, got %#v", node)
	}
	if list.Expr == nil {
		t.Fatalf("Expected a Type")
	}
}

func TestCheckListType(t *testing.T) {
	// []number
	parser := MakeParser(nil)
	expr := &ListTypeExpression{
		Expr: &Literal{Token: token{kind: NumberKeyword}},
	}
	expr.typeCheck(parser)

	if len(parser.errors) != 0 {
		t.Fatalf("Expected no errors, got %#v", parser.errors)
	}
	ty, ok := expr.Type().(Type)
	if !ok {
		t.Fatal("Expected type")
	}
	list, ok := ty.Value.(List)
	if !ok {
		t.Fatal("Expected list type")
	}
	if _, ok := list.Element.(Number); !ok {
		t.Fatal("Expected number list type")
	}
}

func TestCheckListTypeNoValue(t *testing.T) {
	// []42
	parser := MakeParser(nil)
	expr := &ListTypeExpression{
		Bracketed: &BracketedExpression{},
		Expr:      &Literal{Token: literal{kind: NumberLiteral, value: "42"}},
	}
	expr.typeCheck(parser)

	if len(parser.errors) != 1 {
		t.Fatalf("Expected 1 error, got %#v", parser.errors)
	}
	ty, ok := expr.Type().(Type)
	if !ok {
		t.Fatal("Expected type")
	}
	list, ok := ty.Value.(List)
	if !ok {
		t.Fatal("Expected list type")
	}
	if _, ok := list.Element.(Unknown); !ok {
		t.Fatal("Expected unknown list type")
	}
}

func TestNestedListTypeExpression(t *testing.T) {
	tokenizer := testTokenizer{tokens: []Token{
		token{kind: LeftBracket},
		token{kind: RightBracket},
		token{kind: LeftBracket},
		token{kind: RightBracket},
		token{kind: NumberKeyword},
	}}
	parser := MakeParser(&tokenizer)
	node := ParseExpression(parser)

	if len(parser.errors) != 0 {
		t.Fatalf("Expected no errors, got %+v: %#v", len(parser.errors), parser.errors)
	}

	list, ok := node.(*ListTypeExpression)
	if !ok {
		t.Fatalf("Expected ListExpression, got %#v", node)
	}
	if _, ok := list.Expr.(*ListTypeExpression); !ok {
		t.Fatalf("Expected a nested ListTypeExpression, got %#v", list.Type())
	}
	if list.Expr == nil {
		t.Fatalf("Expected a Type")
	}
}

func TestListTypeExpressionWithBracketed(t *testing.T) {
	parser := MakeParser(&testTokenizer{tokens: []Token{
		token{kind: LeftBracket},
		token{kind: NumberKeyword},
		token{kind: RightBracket},
		token{kind: NumberKeyword},
	}})
	parser.parseExpression()

	if len(parser.errors) != 1 {
		t.Fatalf("Expected 1 error, got %#v", parser.errors)
	}
}
