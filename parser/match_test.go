package parser

import "testing"

func TestMatch(t *testing.T) {
	tokenizer := testTokenizer{tokens: []Token{
		token{kind: MATCH_KW},
		literal{kind: IDENTIFIER, value: "option"},
		token{kind: LBRACE},
		token{kind: EOL},

		token{kind: CASE_KW},
		literal{kind: IDENTIFIER, value: "Some"},
		token{kind: LPAREN},
		literal{kind: IDENTIFIER, value: "s"},
		token{kind: RPAREN},
		token{kind: COLON},
		token{kind: EOL},

		token{kind: RETURN_KW},
		literal{kind: IDENTIFIER, value: "s"},
		token{kind: EOL},

		token{kind: RBRACE},
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
