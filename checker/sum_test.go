package checker

import (
	"testing"

	"github.com/bmelicque/test-parser/parser"
)

func TestSumTypeDefinition(t *testing.T) {
	checker := MakeChecker()
	sum := checker.checkSumType(parser.SumType{
		Members: []parser.Node{
			parser.TypedExpression{
				Expr:   parser.TokenExpression{Token: testToken{parser.Name, "Some", parser.Loc{}}},
				Typing: parser.TokenExpression{Token: testToken{parser.NumberKeyword, "number", parser.Loc{}}},
			},
			parser.TokenExpression{Token: testToken{parser.Name, "None", parser.Loc{}}},
		},
	})

	if len(checker.errors) != 0 {
		t.Fatalf("Expected no errors, got %#v", checker.errors)
	}

	if len(sum.Members) != 2 {
		t.Fatalf("Expected 2 members, got %v", len(sum.Members))
	}

	member := sum.Members[0]
	if _, ok := member.Typing.(Literal); !ok {
		t.Fatalf("Expected literal type 'number', got %#v", member.Typing)
	}

	ty, ok := sum.Type().(Type)
	if !ok {
		t.Fatalf("Expected typing, got %#v", sum.Type())
	}
	s, ok := ty.Value.(Sum)
	if !ok {
		t.Fatalf("Expected sum type, got %#v", ty.Value)
	}
	some := s.Members["Some"]
	if some.Kind() != NUMBER {
		t.Fatalf("Expected Some constructor to be a number")
	}
}

func TestSumTypeLength(t *testing.T) {
	checker := MakeChecker()
	checker.checkSumType(parser.SumType{
		Members: []parser.Node{
			parser.TokenExpression{Token: testToken{parser.Name, "Alone", parser.Loc{}}},
		},
	})

	if len(checker.errors) != 1 {
		t.Fatalf("Expected one error, got %#v", checker.errors)
	}
}
