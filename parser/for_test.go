package parser

import (
	"strings"
	"testing"
)

func TestParseForEmptyExpression(t *testing.T) {
	parser := MakeParser(strings.NewReader("for { 42 }"))
	parser.parseForExpression()

	if len(parser.errors) != 0 {
		t.Fatalf("Expected no errors, got %#v", parser.errors)
	}
}

func TestParseForExpression(t *testing.T) {
	parser := MakeParser(strings.NewReader("for true { 42 }"))
	parser.parseForExpression()

	if len(parser.errors) != 0 {
		t.Fatalf("Expected no errors, got %#v", parser.errors)
	}
}

func TestParseForInRangeExpression(t *testing.T) {
	parser := MakeParser(strings.NewReader("for i in 0..=4 { i }"))
	expr := parser.parseForExpression()

	if len(parser.errors) != 0 {
		t.Fatalf("Expected no errors, got %#v", parser.errors)
	}

	binary, ok := expr.Expr.(*BinaryExpression)
	if !ok || binary.Operator.Kind() != InKeyword {
		t.Fatalf("Expected 'in' expression, got:\n%#v", expr.Expr)
	}
}

func TestParseForInExpression(t *testing.T) {
	parser := MakeParser(strings.NewReader("for el in array { 42 }"))
	parser.parseForExpression()

	if len(parser.errors) != 0 {
		t.Fatalf("Expected no errors, got %#v", parser.errors)
	}
}

func TestParseForInTupleExpression(t *testing.T) {
	parser := MakeParser(strings.NewReader("for el, i in array { 42 }"))
	parser.parseForExpression()

	if len(parser.errors) != 0 {
		t.Fatalf("Expected no errors, got %#v", parser.errors)
	}
}

func TestParseForInTupleTooMany(t *testing.T) {
	parser := MakeParser(strings.NewReader("for el, i, extra in array { 42 }"))
	parser.parseForExpression()

	if len(parser.errors) != 1 {
		t.Fatalf("Expected 1 error, got %#v", parser.errors)
	}
}

func TestParseForInExpressionMissingIdentifier(t *testing.T) {
	parser := MakeParser(strings.NewReader("for in array { 42 }"))
	parser.parseForExpression()

	if len(parser.errors) != 1 {
		t.Fatalf("Expected 1 error, got %v:\n %#v", len(parser.errors), parser.errors)
	}
}

func TestCheckForExpressionType(t *testing.T) {
	expr := &ForExpression{
		Keyword: token{kind: ForKeyword},
		Body: &Block{Statements: []Node{
			&Exit{
				Operator: token{kind: BreakKeyword},
				Value:    &Literal{literal{kind: NumberLiteral, value: "42"}},
			},
		}},
	}

	parser := MakeParser(nil)
	expr.typeCheck(parser)

	if len(parser.errors) != 0 {
		t.Fatalf("Expected no errors, got %#v", parser.errors)
	}

	alias, ok := expr.Type().(TypeAlias)
	if !ok || alias.Name != "?" {
		t.Fatalf("Expected ?number, got %v", expr.Type().Text())
	}

	if _, ok := getSomeType(alias.Ref.(Sum)).(Number); !ok {
		t.Fatalf("Expected number type, got %#v", expr.Type())
	}
}

func TestCheckForExpressionCondition(t *testing.T) {
	parser := MakeParser(nil)
	expr := &ForExpression{
		Keyword: token{kind: ForKeyword},
		Expr:    &Literal{literal{kind: BooleanLiteral, value: "true"}},
		Body:    &Block{},
	}
	expr.typeCheck(parser)
	if len(parser.errors) != 0 {
		t.Fatalf("Expected no errors, got %#v", parser.errors)
	}

	expr.Expr = &Literal{literal{kind: NumberLiteral, value: "42"}}
	expr.typeCheck(parser)
	if len(parser.errors) != 1 {
		t.Fatalf("Expected 1 error, got %#v", parser.errors)
	}
}

func TestCheckForInList(t *testing.T) {
	parser := MakeParser(nil)
	parser.scope.Add("list", Loc{}, List{Number{}})
	expr := &ForExpression{
		Keyword: token{kind: ForKeyword},
		Expr: &BinaryExpression{
			Left:     &Identifier{Token: &literal{kind: Name, value: "el"}},
			Right:    &Identifier{Token: &literal{kind: Name, value: "list"}},
			Operator: token{kind: InKeyword},
		},
		Body: &Block{Statements: []Node{
			&Identifier{Token: &literal{kind: Name, value: "el"}},
		}},
	}
	expr.typeCheck(parser)
	if len(parser.errors) != 0 {
		t.Fatalf("Expected no errors, got %#v", parser.errors)
	}

	v, ok := expr.Body.scope.Find("el")
	if !ok {
		t.Fatalf("Expected to find 'el' variable in scope")
	}
	if _, ok := v.Typing.(Number); !ok {
		t.Fatalf("Expected 'el' to be a number")
	}
}

func TestCheckForInWithTuple(t *testing.T) {
	parser := MakeParser(nil)
	parser.scope.Add("list", Loc{}, List{Number{}})
	expr := &ForExpression{
		Keyword: token{kind: ForKeyword},
		Expr: &BinaryExpression{
			Left: &TupleExpression{Elements: []Expression{
				&Identifier{Token: &literal{kind: Name, value: "el"}},
				&Identifier{Token: &literal{kind: Name, value: "i"}},
			}},
			Right:    &Identifier{Token: &literal{kind: Name, value: "list"}},
			Operator: token{kind: InKeyword},
		},
		Body: &Block{Statements: []Node{
			&Identifier{Token: &literal{kind: Name, value: "el"}},
			&Identifier{Token: &literal{kind: Name, value: "i"}},
		}},
	}
	expr.typeCheck(parser)
	if len(parser.errors) != 0 {
		t.Fatalf("Expected no errors, got %#v", parser.errors)
	}

	v, ok := expr.Body.scope.Find("el")
	if !ok {
		t.Fatalf("Expected to find 'el' variable in scope")
	}
	if _, ok := v.Typing.(Number); !ok {
		t.Fatalf("Expected 'el' to be a number")
	}

	i, ok := expr.Body.scope.Find("i")
	if !ok {
		t.Fatalf("Expected to find 'i' variable in scope")
	}
	if _, ok := i.Typing.(Number); !ok {
		t.Fatalf("Expected 'i' to be a number")
	}
}

func TestCheckForInBadType(t *testing.T) {
	parser := MakeParser(nil)
	parser.scope.Add("bad", Loc{}, Ref{Number{}})
	expr := &ForExpression{
		Keyword: token{kind: ForKeyword},
		Expr: &BinaryExpression{
			Left:     &Identifier{Token: &literal{kind: Name, value: "el"}},
			Right:    &Identifier{Token: &literal{kind: Name, value: "bad"}},
			Operator: token{kind: InKeyword},
		},
		Body: &Block{Statements: []Node{
			&Identifier{Token: &literal{kind: Name, value: "el"}},
		}},
	}
	expr.typeCheck(parser)
	if len(parser.errors) != 1 {
		t.Fatalf("Expected 1 error, got %#v", parser.errors)
	}

	v, ok := expr.Body.scope.Find("el")
	if !ok {
		t.Fatalf("Expected to find 'el' variable in scope")
	}
	if v.Typing != (Invalid{}) {
		t.Fatalf("Expected 'el' to be invalid, got '%v'", v.Typing.Text())
	}
}
