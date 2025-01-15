package parser

import (
	"reflect"
	"strings"
	"testing"
)

func TestParseCatchExpression(t *testing.T) {
	tests := []struct {
		name        string
		source      string
		errorCount  int
		expectedLoc Loc
	}{
		{
			name:        "Valid catch expression produces no errors",
			source:      "result catch err {}",
			errorCount:  0,
			expectedLoc: Loc{Position{1, 1}, Position{1, 20}},
		},
		{
			name:        "Catch expression with no identifiers are valid",
			source:      "result catch {}",
			errorCount:  0,
			expectedLoc: Loc{Position{1, 1}, Position{1, 16}},
		},
		{
			name:        "Invalid identifiers produce one error",
			source:      "result catch number {}",
			errorCount:  1,
			expectedLoc: Loc{Position{1, 1}, Position{1, 23}},
		},
		{
			name:        "Invalid tokens produce one error",
			source:      "result catch err err {}",
			errorCount:  1,
			expectedLoc: Loc{Position{1, 1}, Position{1, 24}},
		},
		{
			name:        "Missing body produces one error",
			source:      "result catch err",
			errorCount:  1,
			expectedLoc: Loc{Position{1, 1}, Position{1, 17}},
		},
		{
			name:        "Missing body & identifier produces one error",
			source:      "result catch",
			errorCount:  1,
			expectedLoc: Loc{Position{1, 1}, Position{1, 13}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := MakeParser(strings.NewReader(tt.source))
			expr := parser.parseCatchExpression()
			if got := len(parser.errors); got != tt.errorCount {
				t.Errorf("Got %v errors, want %v\n", got, tt.errorCount)
				for _, e := range parser.errors {
					t.Log(e.Text())
				}
			}
			if _, ok := expr.(*CatchExpression); !ok {
				t.Errorf("Expected *CatchExpression, got %#v", reflect.TypeOf(expr))
			}
			if expr.Loc() != tt.expectedLoc {
				t.Errorf("Got loc %v, want %v", expr.Loc(), tt.expectedLoc)
			}
		})
	}
}

func TestCheckCatchExpression(t *testing.T) {
	parser := MakeParser(nil)
	parser.scope.Add(
		"result",
		Loc{},
		makeResultType(Number{}, String{}),
	)
	expr := &CatchExpression{
		Left:       &Identifier{Token: literal{kind: Name, value: "result"}},
		Keyword:    token{kind: CatchKeyword},
		Identifier: &Identifier{Token: literal{kind: Name, value: "err"}},
		Body: &Block{Statements: []Node{
			&Identifier{Token: literal{kind: Name, value: "err"}},
			&Literal{literal{kind: NumberLiteral, value: "0"}},
		}},
	}
	expr.typeCheck(parser)

	if len(parser.errors) > 0 {
		t.Fatalf("Expected no errors, got %#v", parser.errors)
	}
	if _, ok := expr.Type().(Number); !ok {
		t.Fatalf("Expected number")
	}
}

func TestCheckCatchExpressionNotResult(t *testing.T) {
	parser := MakeParser(nil)
	expr := &CatchExpression{
		Left:       &Literal{literal{kind: NumberLiteral, value: "42"}},
		Keyword:    token{kind: CatchKeyword},
		Identifier: &Identifier{Token: literal{kind: Name, value: "err"}},
		Body: &Block{Statements: []Node{
			&Identifier{Token: literal{kind: Name, value: "err"}},
			&Literal{literal{kind: NumberLiteral, value: "0"}},
		}},
	}
	expr.typeCheck(parser)

	if len(parser.errors) != 1 {
		t.Fatalf("Expected 1 error, got %v: %#v", len(parser.errors), parser.errors)
	}
	if _, ok := expr.Type().(Number); !ok {
		t.Fatalf("Expected number")
	}
}

func TestCheckCatchExpressionBlockNotMatching(t *testing.T) {
	parser := MakeParser(nil)
	parser.scope.Add(
		"result",
		Loc{},
		makeResultType(Number{}, String{}),
	)
	expr := &CatchExpression{
		Left:       &Identifier{Token: literal{kind: Name, value: "result"}},
		Keyword:    token{kind: CatchKeyword},
		Identifier: &Identifier{Token: literal{kind: Name, value: "err"}},
		Body: &Block{Statements: []Node{
			&Identifier{Token: literal{kind: Name, value: "err"}},
		}},
	}
	expr.typeCheck(parser)

	if len(parser.errors) != 1 {
		t.Fatalf("Expected 1 error, got %v: %#v", len(parser.errors), parser.errors)
	}
	if _, ok := expr.Body.Type().(String); !ok {
		t.Fatalf("Expected string, got %#v", expr.Type())
	}
}
