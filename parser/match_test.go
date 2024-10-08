package parser

import "testing"

func TestMatch(t *testing.T) {
	tokenizer := testTokenizer{tokens: []Token{
		token{kind: MatchKeyword},
		literal{kind: Name, value: "option"},
		token{kind: LeftBrace},
		token{kind: EOL},

		token{kind: CaseKeyword},
		literal{kind: Name, value: "Some"},
		token{kind: LeftParenthesis},
		literal{kind: Name, value: "s"},
		token{kind: RightParenthesis},
		token{kind: Colon},
		token{kind: EOL},

		token{kind: ReturnKeyword},
		literal{kind: Name, value: "s"},
		token{kind: EOL},

		token{kind: RightBrace},
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
