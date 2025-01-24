package parser

import (
	"strings"
	"testing"
)

func TestParseMatchCase(t *testing.T) {
	tests := []struct {
		name      string
		source    string
		wantError bool
	}{
		{
			name:      "case enum",
			source:    "case None:",
			wantError: false,
		},
		{
			name:      "case constructor",
			source:    "case Some{s}:",
			wantError: false,
		},
		{
			name:      "case constructor with statement",
			source:    "case Some{s}: s",
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := MakeParser(strings.NewReader(tt.source))
			parseMatchCase(parser)
			if tt.wantError && len(parser.errors) == 0 {
				t.Error("Got no errors, want one\n")
			}
			if !tt.wantError && len(parser.errors) > 0 {
				t.Error("Got one error, want none\n")
				t.Log(parser.errors[0].Text())
			}
		})
	}
}

func TestMatch(t *testing.T) {
	str := "match option {\n"
	str += "case Some{s}:\n"
	str += "    return s\n"
	str += "}"
	parser := MakeParser(strings.NewReader(str))
	node := parser.parseMatchExpression()
	statement, ok := node.(*MatchExpression)
	if !ok {
		t.Fatalf("Expected match expression, got %#v", node)
	}

	if len(statement.Cases) != 1 {
		t.Fatalf("Expected 1 case, got %#v", statement.Cases)
	}
}
