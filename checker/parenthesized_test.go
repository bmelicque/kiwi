package checker

import (
	"testing"

	"github.com/bmelicque/test-parser/parser"
	"github.com/bmelicque/test-parser/tokenizer"
)

func TestParam(t *testing.T) {
	checker := MakeChecker()
	expr := checker.checkParam(parser.TypedExpression{
		Expr:   parser.TokenExpression{Token: testToken{kind: tokenizer.IDENTIFIER, value: "name"}},
		Typing: parser.TokenExpression{Token: testToken{kind: tokenizer.NUM_KW}},
		Colon:  false,
	})

	if len(checker.errors) != 0 {
		t.Fatalf("Expected no errors, got %#v", checker.errors)
	}
	if expr.Identifier.Text() != "name" {
		t.Fatalf("Expected name 'name', got '%v'", expr.Identifier.Text())
	}
	if expr.Complement.Type().Kind() != TYPE {
		t.Fatal("Expected type")
	}
}

func TestParamBadType(t *testing.T) {
	checker := MakeChecker()
	// name 42
	checker.checkParam(parser.TypedExpression{
		Expr:   parser.TokenExpression{Token: testToken{kind: tokenizer.IDENTIFIER, value: "name"}},
		Typing: parser.TokenExpression{Token: testToken{kind: tokenizer.NUMBER, value: "42"}},
		Colon:  false,
	})

	if len(checker.errors) != 1 {
		t.Fatalf("Expected 1 error, got %#v", checker.errors)
	}
}

func TestTypeParam(t *testing.T) {
	checker := MakeChecker()
	expr := checker.checkParam(parser.TypedExpression{
		Expr:  parser.TokenExpression{Token: testToken{kind: tokenizer.IDENTIFIER, value: "Name"}},
		Colon: false,
	})

	if len(checker.errors) != 0 {
		t.Fatalf("Expected no errors, got %#v", checker.errors)
	}
	if expr.Identifier.Text() != "Name" {
		t.Fatalf("Expected name 'Name', got '%v'", expr.Identifier.Text())
	}
}

// accept only type identifiers for type params
func TestTypeParamBadName(t *testing.T) {
	checker := MakeChecker()
	checker.checkTypeParam(parser.TypedExpression{
		Expr:  parser.TokenExpression{Token: testToken{kind: tokenizer.IDENTIFIER, value: "name"}},
		Colon: false,
	})

	if len(checker.errors) != 1 {
		t.Fatalf("Expected 1 error, got %#v", checker.errors)
	}
}

func TestParams(t *testing.T) {
	checker := MakeChecker()
	expr := checker.checkParams(
		parser.ParenthesizedExpression{
			Expr: parser.TypedExpression{
				Expr:   parser.TokenExpression{Token: testToken{kind: tokenizer.IDENTIFIER, value: "name"}},
				Typing: parser.TokenExpression{Token: testToken{kind: tokenizer.NUM_KW}},
				Colon:  false,
			},
		},
	)

	if len(checker.errors) != 0 {
		t.Fatalf("Expected no errors, got %#v", checker.errors)
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

func TestSimpleObject(t *testing.T) {
	checker := MakeChecker()
	expr := checker.checkParenthesizedExpression(
		parser.ParenthesizedExpression{
			Expr: parser.TypedExpression{
				Expr:   parser.TokenExpression{Token: testToken{kind: tokenizer.IDENTIFIER, value: "name"}},
				Typing: parser.TokenExpression{Token: testToken{kind: tokenizer.NUM_KW}},
				Colon:  false,
			},
		},
	)

	if len(checker.errors) != 0 {
		t.Fatalf("Expected no errors, got %#v", checker.errors)
	}

	ty, ok := expr.Type().(Type)
	if !ok {
		t.Fatalf("Expected a typing, got %#v", expr.Type())
	}
	if _, ok := ty.Value.(Object); !ok {
		t.Fatalf("Expected an object typing, got %#v", expr.Type())
	}
}
