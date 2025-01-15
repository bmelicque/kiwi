package parser

import (
	"strings"
	"testing"
)

func TestBreak(t *testing.T) {
	parser := MakeParser(strings.NewReader("break true"))
	parser.pushScope(NewScope(LoopScope))
	exit := parser.parseExit()

	if len(parser.errors) != 0 {
		t.Fatalf("Expected no errors, got %#v", parser.errors)
	}
	if exit.Operator.Kind() != BreakKeyword {
		t.Fatal("Expected 'break' keyword")
	}
}

func TestBreakOutsideLoop(t *testing.T) {
	parser := MakeParser(strings.NewReader("break true"))
	parser.parseExit()

	if len(parser.errors) != 1 {
		t.Fatalf("Expected 1 error, got %#v", parser.errors)
	}
}

func TestContinue(t *testing.T) {
	parser := MakeParser(strings.NewReader("continue"))
	parser.pushScope(NewScope(LoopScope))
	exit := parser.parseExit()

	if len(parser.errors) != 0 {
		t.Fatalf("Expected no errors, got %#v", parser.errors)
	}
	if exit.Operator.Kind() != ContinueKeyword {
		t.Fatal("Expected 'continue' keyword")
	}
}

func TestContinueOutsideLoop(t *testing.T) {
	parser := MakeParser(strings.NewReader("continue"))
	parser.parseExit()

	if len(parser.errors) != 1 {
		t.Fatalf("Expected 1 error, got %#v", parser.errors)
	}
}

func TestReturn(t *testing.T) {
	parser := MakeParser(strings.NewReader("return true"))
	parser.pushScope(NewScope(FunctionScope))
	exit := parser.parseExit()

	if len(parser.errors) != 0 {
		t.Fatalf("Expected no errors, got %#v", parser.errors)
	}
	if exit.Operator.Kind() != ReturnKeyword {
		t.Fatal("Expected 'return' keyword")
	}
}

func TestReturnOutsideFunction(t *testing.T) {
	parser := MakeParser(strings.NewReader("return true"))
	parser.parseExit()

	if len(parser.errors) != 1 {
		t.Fatalf("Expected 1 error, got %#v", parser.errors)
	}
}

func TestThrow(t *testing.T) {
	parser := MakeParser(strings.NewReader("throw true"))
	parser.pushScope(NewScope(FunctionScope))
	exit := parser.parseExit()

	if len(parser.errors) != 0 {
		t.Fatalf("Expected no errors, got %#v", parser.errors)
	}
	if exit.Operator.Kind() != ThrowKeyword {
		t.Fatal("Expected 'throw' keyword")
	}
}

func TestThrowOutsideFunction(t *testing.T) {
	parser := MakeParser(strings.NewReader("throw true"))
	parser.parseExit()

	if len(parser.errors) != 1 {
		t.Fatalf("Expected 1 error, got %#v", parser.errors)
	}
}

func TestIsExiting(t *testing.T) {
	tests := []struct {
		name     string
		node     Node
		expected bool
	}{
		{
			name:     "Exit node returns true",
			node:     &Exit{},
			expected: true,
		},
		{
			name:     "Empty block returns false",
			node:     &Block{Statements: []Node{}},
			expected: false,
		},
		{
			name:     "Block with non-exiting statements returns false",
			node:     &Block{Statements: []Node{&Block{}, &Block{}}},
			expected: false,
		},
		{
			name:     "Block with one exiting statement returns true",
			node:     &Block{Statements: []Node{&Block{}, &Exit{}, &Block{}}},
			expected: true,
		},
		{
			name: "If expression with both branches exiting returns true",
			node: &IfExpression{
				Body:      &Block{Statements: []Node{&Exit{}}},
				Alternate: &Block{Statements: []Node{&Exit{}}},
			},
			expected: true,
		},
		{
			name: "If expression with only consequence exiting returns false",
			node: &IfExpression{
				Body:      &Block{Statements: []Node{&Exit{}}},
				Alternate: &Block{Statements: []Node{}}},
			expected: false,
		},
		{
			name: "If expression with only alternate exiting returns false",
			node: &IfExpression{
				Body:      &Block{Statements: []Node{}},
				Alternate: &Block{Statements: []Node{&Exit{}}},
			},
			expected: false,
		},
		{
			name: "If expression with neither branch exiting returns false",
			node: &IfExpression{
				Body:      &Block{Statements: []Node{}},
				Alternate: &Block{Statements: []Node{}},
			},
			expected: false,
		},
		{
			name:     "Other node type returns false",
			node:     &Identifier{},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isExiting(tt.node); got != tt.expected {
				t.Errorf("isExiting() = %v, want %v", got, tt.expected)
			}
		})
	}
}
