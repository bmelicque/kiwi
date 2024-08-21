package checker

import (
	"testing"

	"github.com/bmelicque/test-parser/parser"
	"github.com/bmelicque/test-parser/tokenizer"
)

func TestObjectDescription(t *testing.T) {
	checker := MakeChecker()
	object := checker.checkObjectDefinition(parser.ObjectDefinition{
		Members: []parser.Node{
			parser.TypedExpression{
				Expr:   parser.TokenExpression{Token: testToken{tokenizer.IDENTIFIER, "n", tokenizer.Loc{}}},
				Typing: parser.TokenExpression{Token: testToken{tokenizer.NUM_KW, "number", tokenizer.Loc{}}},
			},
			parser.TypedExpression{
				Expr:   parser.TokenExpression{Token: testToken{tokenizer.IDENTIFIER, "s", tokenizer.Loc{}}},
				Typing: parser.TokenExpression{Token: testToken{tokenizer.NUM_KW, "string", tokenizer.Loc{}}},
			},
		},
	})

	if len(checker.errors) != 0 {
		t.Fatalf("Expected no errors, got %#v", checker.errors)
	}

	if len(object.Members) != 2 {
		t.Fatalf("Expected 2 members, got %v", len(object.Members))
	}
}

func TestObjectDescriptionDuplicates(t *testing.T) {
	checker := MakeChecker()
	object := checker.checkObjectDefinition(parser.ObjectDefinition{
		Members: []parser.Node{
			parser.TypedExpression{
				Expr:   parser.TokenExpression{Token: testToken{tokenizer.IDENTIFIER, "n", tokenizer.Loc{}}},
				Typing: parser.TokenExpression{Token: testToken{tokenizer.NUM_KW, "number", tokenizer.Loc{}}},
			},
			parser.TypedExpression{
				Expr:   parser.TokenExpression{Token: testToken{tokenizer.IDENTIFIER, "n", tokenizer.Loc{}}},
				Typing: parser.TokenExpression{Token: testToken{tokenizer.NUM_KW, "number", tokenizer.Loc{}}},
			},
		},
	})

	if len(checker.errors) != 2 {
		// One error on each duplicated member
		t.Fatalf("Expected 2 errors, got %#v", checker.errors)
	}

	if len(object.Members) != 2 {
		t.Fatalf("Expected 2 members, got %v", len(object.Members))
	}
}

func TestObjectDescriptionColons(t *testing.T) {
	checker := MakeChecker()
	checker.checkObjectDefinition(parser.ObjectDefinition{
		Members: []parser.Node{
			parser.TypedExpression{
				Expr:   parser.TokenExpression{Token: testToken{tokenizer.IDENTIFIER, "n", tokenizer.Loc{}}},
				Typing: parser.TokenExpression{Token: testToken{tokenizer.NUM_KW, "number", tokenizer.Loc{}}},
				Colon:  true,
			},
		},
	})

	if len(checker.errors) != 1 {
		t.Fatalf("Expected 1 error, got %v: %#v", len(checker.errors), checker.errors)
	}
}
