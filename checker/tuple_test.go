package checker

import (
	"testing"

	"github.com/bmelicque/test-parser/parser"
)

func TestObjectType(t *testing.T) {
	checker := MakeChecker()
	//	n number, s string
	tuple := checker.checkTuple(parser.TupleExpression{
		Elements: []parser.Node{
			parser.TypedExpression{
				Expr:   parser.TokenExpression{Token: testToken{kind: parser.Name, value: "n"}},
				Typing: parser.TokenExpression{Token: testToken{kind: parser.NumberKeyword}},
			},
			parser.TypedExpression{
				Expr:   parser.TokenExpression{Token: testToken{kind: parser.Name, value: "s"}},
				Typing: parser.TokenExpression{Token: testToken{kind: parser.StringKeyword}},
			},
		},
	})

	if len(checker.errors) != 0 {
		t.Fatalf("Expected no errors, got %#v", checker.errors)
	}

	if len(tuple.Elements) != 2 {
		t.Fatalf("Expected 2 members, got %v", len(tuple.Elements))
	}

	ty, ok := tuple.Type().(Type)
	if !ok {
		t.Fatalf("Expected object type, got %#v", tuple.Type())
	}
	object, ok := ty.Value.(Object)
	if !ok {
		t.Fatalf("Expected object type, got %#v", tuple.Type())
	}
	n, ok := object.Members["n"]
	if !ok {
		t.Fatal("Couldn't find key 'n'")
	}
	if n.Kind() != NUMBER {
		t.Fatal("Expected 'n' to be a number")
	}
}

func TestObjectTypeDuplicates(t *testing.T) {
	checker := MakeChecker()
	// n number, n number
	checker.checkTuple(parser.TupleExpression{
		Elements: []parser.Node{
			parser.TypedExpression{
				Expr:   parser.TokenExpression{Token: testToken{kind: parser.Name, value: "n"}},
				Typing: parser.TokenExpression{Token: testToken{kind: parser.NumberKeyword}},
			},
			parser.TypedExpression{
				Expr:   parser.TokenExpression{Token: testToken{kind: parser.Name, value: "n"}},
				Typing: parser.TokenExpression{Token: testToken{kind: parser.NumberKeyword}},
			},
		},
	})

	if len(checker.errors) != 2 {
		// One error on each duplicated member
		t.Fatalf("Expected 2 errors, got %#v", checker.errors)
	}
}

func TestTupleConsistency(t *testing.T) {
	checker := MakeChecker()
	// n number, 42
	checker.checkTuple(parser.TupleExpression{
		Elements: []parser.Node{
			parser.TypedExpression{
				Expr:   parser.TokenExpression{Token: testToken{kind: parser.Name, value: "n"}},
				Typing: parser.TokenExpression{Token: testToken{kind: parser.NumberKeyword}},
			},
			parser.TokenExpression{Token: testToken{kind: parser.NumberLiteral, value: "42"}},
		},
	})

	if len(checker.errors) != 1 {
		t.Fatalf("Expected 1 error, got %#v", checker.errors)
	}
}
