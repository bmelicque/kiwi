package checker

import (
	"testing"

	"github.com/bmelicque/test-parser/parser"
	"github.com/bmelicque/test-parser/tokenizer"
)

func TestObjectType(t *testing.T) {
	checker := MakeChecker()
	tuple := checker.checkTuple(parser.TupleExpression{
		Elements: []parser.Node{
			parser.TypedExpression{
				Expr:   parser.TokenExpression{Token: testToken{kind: tokenizer.IDENTIFIER, value: "n"}},
				Typing: parser.TokenExpression{Token: testToken{kind: tokenizer.NUM_KW}},
			},
			parser.TypedExpression{
				Expr:   parser.TokenExpression{Token: testToken{kind: tokenizer.IDENTIFIER, value: "s"}},
				Typing: parser.TokenExpression{Token: testToken{kind: tokenizer.STR_KW}},
			},
		},
	})

	if len(checker.errors) != 0 {
		t.Fatalf("Expected no errors, got %#v", checker.errors)
	}

	if len(tuple.Elements) != 2 {
		t.Fatalf("Expected 2 members, got %v", len(tuple.Elements))
	}

	if _, ok := tuple.Type().(Type); !ok {
		t.Fatalf("Expected object type, got %#v", tuple.Type())
	}
}

func TestObjectTypeWithIdentifiers(t *testing.T) {
	checker := MakeChecker()
	checker.scope.Add("Type", tokenizer.Loc{}, Type{TypeAlias{Name: "Type", Ref: Object{map[string]ExpressionType{}}}})
	tuple := checker.checkTuple(parser.TupleExpression{
		Elements: []parser.Node{
			parser.TypedExpression{
				Expr:   parser.TokenExpression{Token: testToken{kind: tokenizer.IDENTIFIER, value: "value"}},
				Typing: parser.TokenExpression{Token: testToken{kind: tokenizer.IDENTIFIER, value: "Type"}},
			},
		},
	})

	if len(checker.errors) != 0 {
		t.Fatalf("Expected no errors, got %#v", checker.errors)
	}

	param := tuple.Elements[0].(Param)
	if _, ok := param.Complement.(Identifier); !ok {
		t.Fatalf("Expected type 'Type', got %#v", param.Complement)
	}
}

func TestObjectTypeWithDuplicates(t *testing.T) {
	checker := MakeChecker()
	checker.checkTuple(parser.TupleExpression{
		Elements: []parser.Node{
			parser.TypedExpression{
				Expr:   parser.TokenExpression{Token: testToken{kind: tokenizer.IDENTIFIER, value: "n"}},
				Typing: parser.TokenExpression{Token: testToken{kind: tokenizer.NUM_KW}},
			},
			parser.TypedExpression{
				Expr:   parser.TokenExpression{Token: testToken{kind: tokenizer.IDENTIFIER, value: "n"}},
				Typing: parser.TokenExpression{Token: testToken{kind: tokenizer.NUM_KW}},
			},
		},
	})

	if len(checker.errors) != 2 {
		// One error on each duplicated member
		t.Fatalf("Expected 2 errors, got %#v", checker.errors)
	}
}
