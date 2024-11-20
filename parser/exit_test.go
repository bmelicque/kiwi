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
