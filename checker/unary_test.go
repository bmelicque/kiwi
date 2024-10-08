package checker

import (
	"testing"

	"github.com/bmelicque/test-parser/parser"
)

func TestUnaryExpression(t *testing.T) {
	checker := MakeChecker()
	expr := checker.checkUnaryExpression(parser.UnaryExpression{
		Operator: testToken{kind: parser.QUESTION_MARK},
		Operand:  parser.TokenExpression{Token: testToken{kind: parser.NUM_KW}},
	})

	if len(checker.errors) != 0 {
		t.Fatalf("Expected no errors, got %#v", checker.errors)
	}
	if expr.Operator.Kind() != parser.QUESTION_MARK {
		t.Fatal("Expected question mark")
	}
	if _, ok := expr.Operand.(Literal); !ok {
		t.Fatal("Expected literal")
	}
}

func TestNestedUnaryExpression(t *testing.T) {
	checker := MakeChecker()
	expr := checker.checkUnaryExpression(parser.UnaryExpression{
		Operator: testToken{kind: parser.QUESTION_MARK},
		Operand: parser.UnaryExpression{
			Operator: testToken{kind: parser.QUESTION_MARK},
			Operand:  parser.TokenExpression{Token: testToken{kind: parser.NUM_KW}},
		},
	})

	if len(checker.errors) != 0 {
		t.Fatalf("Expected no errors, got %#v", checker.errors)
	}
	if _, ok := expr.Operand.(UnaryExpression); !ok {
		t.Fatal("Expected unary expression")
	}
}

func TestNoOptionValue(t *testing.T) {
	checker := MakeChecker()
	checker.checkUnaryExpression(parser.UnaryExpression{
		Operator: testToken{kind: parser.QUESTION_MARK},
		Operand:  parser.TokenExpression{Token: testToken{kind: parser.NUMBER, value: "42"}},
	})

	if len(checker.errors) != 1 {
		t.Fatalf("Expected 1 error, got %#v", checker.errors)
	}
}
