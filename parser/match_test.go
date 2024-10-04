package parser

import (
	"testing"

	"github.com/bmelicque/test-parser/tokenizer"
)

func TestMatch(t *testing.T) {
	tokenizer := testTokenizer{tokens: []tokenizer.Token{
		testToken{kind: tokenizer.MATCH_KW},
		testToken{kind: tokenizer.IDENTIFIER, value: "option"},
		testToken{kind: tokenizer.LBRACE},
		testToken{kind: tokenizer.EOL},

		testToken{kind: tokenizer.CASE_KW},
		testToken{kind: tokenizer.IDENTIFIER, value: "Some"},
		testToken{kind: tokenizer.LPAREN},
		testToken{kind: tokenizer.IDENTIFIER, value: "s"},
		testToken{kind: tokenizer.RPAREN},
		testToken{kind: tokenizer.COLON},
		testToken{kind: tokenizer.EOL},

		testToken{kind: tokenizer.RETURN_KW},
		testToken{kind: tokenizer.IDENTIFIER, value: "s"},
		testToken{kind: tokenizer.EOL},

		testToken{tokenizer.RBRACE, "}", tokenizer.Loc{}},
	}}
	parser := MakeParser(&tokenizer)
	node := parser.parseMatchExpression()
	statement, ok := node.(MatchExpression)
	if !ok {
		t.Fatalf("Expected match statement, got %#v", node)
	}

	if len(statement.Cases) != 1 {
		t.Fatalf("Expected 1 case, got %#v", statement.Cases)
	}
}
