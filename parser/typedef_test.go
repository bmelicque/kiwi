package parser

import (
	"testing"

	"github.com/bmelicque/test-parser/tokenizer"
)

func TestObjectDescriptionSingleLine(t *testing.T) {
	tokenizer := testTokenizer{tokens: []tokenizer.Token{
		testToken{tokenizer.LBRACE, "{", tokenizer.Loc{Start: tokenizer.Position{Line: 1, Col: 1}}},
		testToken{tokenizer.IDENTIFIER, "n", tokenizer.Loc{Start: tokenizer.Position{Line: 1, Col: 2}}},
		testToken{tokenizer.NUM_KW, "number", tokenizer.Loc{Start: tokenizer.Position{Line: 1, Col: 4}}},
		testToken{tokenizer.RBRACE, "}", tokenizer.Loc{Start: tokenizer.Position{Line: 1, Col: 10}}},
	}}
	parser := MakeParser(&tokenizer)
	node := parser.parseObjectDefinition()

	if len(parser.errors) != 0 {
		t.Fatalf("Expected no errors, got %#v", parser.errors)
	}

	object, ok := node.(ObjectDefinition)
	if !ok {
		t.Fatalf("Expected ObjectDefinition, got %#v", node)
		return
	}

	if len(object.Members) != 1 {
		t.Fatalf("Expected 1 member, got %v", len(object.Members))
	}
}

func TestObjectDescription(t *testing.T) {
	tokenizer := testTokenizer{tokens: []tokenizer.Token{
		testToken{tokenizer.LBRACE, "{", tokenizer.Loc{Start: tokenizer.Position{Line: 1, Col: 1}}},
		testToken{tokenizer.EOL, "\n", tokenizer.Loc{Start: tokenizer.Position{Line: 1, Col: 2}}},
		testToken{tokenizer.IDENTIFIER, "n", tokenizer.Loc{Start: tokenizer.Position{Line: 2, Col: 5}}},
		testToken{tokenizer.NUM_KW, "number", tokenizer.Loc{Start: tokenizer.Position{Line: 2, Col: 7}}},
		testToken{tokenizer.COMMA, ",", tokenizer.Loc{Start: tokenizer.Position{Line: 2, Col: 13}}},
		testToken{tokenizer.EOL, "\n", tokenizer.Loc{Start: tokenizer.Position{Line: 2, Col: 14}}},
		testToken{tokenizer.IDENTIFIER, "s", tokenizer.Loc{Start: tokenizer.Position{Line: 3, Col: 5}}},
		testToken{tokenizer.NUM_KW, "string", tokenizer.Loc{Start: tokenizer.Position{Line: 3, Col: 7}}},
		testToken{tokenizer.COMMA, ",", tokenizer.Loc{Start: tokenizer.Position{Line: 3, Col: 13}}},
		testToken{tokenizer.EOL, "\n", tokenizer.Loc{Start: tokenizer.Position{Line: 3, Col: 14}}},
		testToken{tokenizer.RBRACE, "}", tokenizer.Loc{Start: tokenizer.Position{Line: 4, Col: 1}}},
	}}
	parser := MakeParser(&tokenizer)
	node := parser.parseObjectDefinition()

	if len(parser.errors) != 0 {
		t.Fatalf("Expected no errors, got %#v", parser.errors)
	}

	object, ok := node.(ObjectDefinition)
	if !ok {
		t.Fatalf("Expected ObjectDefinition, got %#v", node)
		return
	}

	if len(object.Members) != 2 {
		t.Fatalf("Expected 2 members, got %v", len(object.Members))
	}
}

func TestObjectDescriptionNoColon(t *testing.T) {
	tokenizer := testTokenizer{tokens: []tokenizer.Token{
		testToken{tokenizer.LBRACE, "{", tokenizer.Loc{Start: tokenizer.Position{Line: 1, Col: 1}}},
		testToken{tokenizer.IDENTIFIER, "n", tokenizer.Loc{Start: tokenizer.Position{Line: 1, Col: 2}}},
		testToken{tokenizer.COLON, ":", tokenizer.Loc{Start: tokenizer.Position{Line: 1, Col: 3}}},
		testToken{tokenizer.NUM_KW, "number", tokenizer.Loc{Start: tokenizer.Position{Line: 1, Col: 5}}},
		testToken{tokenizer.RBRACE, "}", tokenizer.Loc{Start: tokenizer.Position{Line: 1, Col: 11}}},
	}}
	parser := MakeParser(&tokenizer)
	node := parser.parseObjectDefinition()

	if len(parser.errors) != 0 {
		t.Fatalf("Expected no errors, got %#v", parser.errors)
	}

	object, ok := node.(ObjectDefinition)
	if !ok {
		t.Fatalf("Expected ObjectDefinition, got %#v", node)
		return
	}

	if len(object.Members) != 1 {
		t.Fatalf("Expected 1 member, got %v", len(object.Members))
		return
	}

	typed, ok := object.Members[0].(TypedExpression)
	if !ok {
		t.Fatalf("Expected TypedExpression, got %#v", object.Members[0])
		return
	}
	if !typed.Colon {
		t.Fatal("Expected ':'")
	}
}
