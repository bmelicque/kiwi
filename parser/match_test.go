package parser

import (
	"strings"
	"testing"
)

func TestParseMatchCase(t *testing.T) {
	tests := []struct {
		name       string
		source     string
		wantError  bool
		isCatchAll bool
	}{
		{
			name:       "enum",
			source:     "None: 42",
			wantError:  false,
			isCatchAll: false,
		},
		{
			name:       "case constructor",
			source:     "s Some: s",
			wantError:  false,
			isCatchAll: false,
		},
		{
			name:       "missing consequent",
			source:     "None:",
			wantError:  true,
			isCatchAll: false,
		},
		{
			name:       "missing consequent after param",
			source:     "s Some:",
			wantError:  true,
			isCatchAll: false,
		},
		{
			name:       "catch-all case",
			source:     "_: 42",
			wantError:  false,
			isCatchAll: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := MakeParser(strings.NewReader(tt.source))
			expr := parseMatchCase(parser)
			if tt.wantError && len(parser.errors) == 0 {
				t.Error("Got no errors, want one\n")
			}
			if !tt.wantError && len(parser.errors) > 0 {
				t.Error("Got one error, want none\n")
				t.Log(parser.errors[0].Text())
			}
			if expr.IsCatchall() != tt.isCatchAll {
				t.Errorf("Got %v, want %v\n", expr.IsCatchall(), tt.isCatchAll)
			}
		})
	}
}

func TestCheckMatchCase(t *testing.T) {
	testCases := []struct {
		consequent   Expression
		expectedType ExpressionType
	}{
		{
			consequent:   &Block{},
			expectedType: Void{},
		},
		{
			consequent:   &Literal{literal{kind: NumberLiteral}},
			expectedType: Number{},
		},
		{
			consequent: &Block{Statements: []Node{
				&Exit{Operator: token{kind: ReturnKeyword}},
			}},
			expectedType: Void{},
		},
	}

	for _, tc := range testCases {
		matchCase := MatchCase{Consequent: tc.consequent}
		if matchCase.Type() != tc.expectedType {
			t.Errorf("Type() = %v, want %v", matchCase.Type(), tc.expectedType)
		}
	}
}

func TestMatch(t *testing.T) {
	str := "match option {\n"
	str += "s Some: s\n"
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
