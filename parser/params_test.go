package parser

import "testing"

func TestParam(t *testing.T) {
	parser := MakeParser(&testTokenizer{})
	expr := parser.getValidatedParam(TypedExpression{
		Expr:   Identifier{Token: literal{kind: Name, value: "name"}},
		Typing: Literal{Token: token{kind: NumberKeyword}},
		Colon:  false,
	})

	if len(parser.errors) != 0 {
		t.Fatalf("Expected no errors, got %#v", parser.errors)
	}
	if expr.Identifier.Text() != "name" {
		t.Fatalf("Expected name 'name', got '%v'", expr.Identifier.Text())
	}
	if expr.Complement.Type().Kind() != TYPE {
		t.Fatal("Expected type")
	}
}

func TestParamBadType(t *testing.T) {
	parser := MakeParser(&testTokenizer{})
	parser.getValidatedParam(TypedExpression{
		Expr:   Identifier{Token: literal{kind: Name, value: "name"}},
		Typing: Literal{Token: literal{kind: NumberLiteral, value: "42"}},
		Colon:  false,
	})

	if len(parser.errors) != 1 {
		t.Fatalf("Expected 1 error, got %#v", parser.errors)
	}
}

func TestTypeParam(t *testing.T) {
	parser := MakeParser(&testTokenizer{})
	expr := parser.getValidatedTypeParam(TypedExpression{
		Expr:  Identifier{Token: literal{kind: Name, value: "Name"}},
		Colon: false,
	})

	if len(parser.errors) != 0 {
		t.Fatalf("Expected no errors, got %#v", parser.errors)
	}
	if expr.Identifier.Text() != "Name" {
		t.Fatalf("Expected name 'Name', got '%v'", expr.Identifier.Text())
	}
}

// accept only type identifiers for type params
func TestTypeParamBadName(t *testing.T) {
	parser := MakeParser(&testTokenizer{})
	parser.getValidatedTypeParam(TypedExpression{
		Expr:  Identifier{Token: literal{kind: Name, value: "name"}},
		Colon: false,
	})

	if len(parser.errors) != 1 {
		t.Fatalf("Expected 1 error, got %#v", parser.errors)
	}
}

func TestParams(t *testing.T) {
	parser := MakeParser(&testTokenizer{})
	expr := parser.getValidatedParams(
		ParenthesizedExpression{
			Expr: TypedExpression{
				Expr:   Identifier{Token: literal{kind: Name, value: "name"}},
				Typing: Literal{Token: token{kind: NumberKeyword}},
				Colon:  false,
			},
		},
	)

	if len(parser.errors) != 0 {
		t.Fatalf("Expected no errors, got %#v", parser.errors)
	}
	if len(expr.Params) != 1 {
		t.Fatalf("Expected exactly 1 param")
	}
	param := expr.Params[0]
	if param.Identifier.Text() != "name" {
		t.Fatalf("Expected name 'name', got '%v'", param.Identifier.Text())
	}
	if param.Complement.Type().Kind() != TYPE {
		t.Fatal("Expected type")
	}
}
