package checker

import (
	"testing"

	"github.com/bmelicque/test-parser/parser"
	"github.com/bmelicque/test-parser/tokenizer"
)

func TestUnaryExpression(t *testing.T) {
	checker := MakeChecker()
	expr := checker.checkUnaryExpression(parser.UnaryExpression{
		Operator: testToken{kind: tokenizer.QUESTION_MARK},
		Operand:  parser.TokenExpression{Token: testToken{kind: tokenizer.NUM_KW}},
	})

	if len(checker.errors) != 0 {
		t.Fatalf("Expected no errors, got %#v", checker.errors)
	}
	if expr.Operator.Kind() != tokenizer.QUESTION_MARK {
		t.Fatal("Expected question mark")
	}
	if _, ok := expr.Operand.(Literal); !ok {
		t.Fatal("Expected literal")
	}
}

func TestNestedUnaryExpression(t *testing.T) {
	checker := MakeChecker()
	expr := checker.checkUnaryExpression(parser.UnaryExpression{
		Operator: testToken{kind: tokenizer.QUESTION_MARK},
		Operand: parser.UnaryExpression{
			Operator: testToken{kind: tokenizer.QUESTION_MARK},
			Operand:  parser.TokenExpression{Token: testToken{kind: tokenizer.NUM_KW}},
		},
	})

	if len(checker.errors) != 0 {
		t.Fatalf("Expected no errors, got %#v", checker.errors)
	}
	if _, ok := expr.Operand.(UnaryExpression); !ok {
		t.Fatal("Expected unary expression")
	}
}
