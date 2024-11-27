package emitter

import (
	"strings"
	"testing"

	"github.com/bmelicque/test-parser/parser"
)

func TestEmptyBlockExpression(t *testing.T) {
	emitter := makeEmitter()
	emitter.emitExpression(&parser.Block{})

	text := emitter.string()
	expected := "undefined"
	if text != expected {
		t.Fatalf("Expected string:\n%v\ngot:\n%v", expected, text)
	}
}

func TestSingleLineBlockExpression(t *testing.T) {
	emitter := makeEmitter()
	emitter.emitExpression(&parser.Block{Statements: []parser.Node{
		&parser.Literal{Token: testToken{kind: parser.NumberLiteral, value: "42"}},
	}})

	text := emitter.string()
	expected := "42"
	if text != expected {
		t.Fatalf("Expected string:\n%v\ngot:\n%v", expected, text)
	}
}

func TestBlockExpression(t *testing.T) {
	emitter := makeEmitter()
	emitter.emitExpression(&parser.Block{Statements: []parser.Node{
		&parser.Literal{Token: testToken{kind: parser.NumberLiteral, value: "42"}},
		&parser.Literal{Token: testToken{kind: parser.NumberLiteral, value: "42"}},
	}})

	text := emitter.string()
	expected := "(\n    42,\n    42,\n)"
	if text != expected {
		t.Fatalf("Expected string:\n%v\ngot:\n%v", expected, text)
	}
}

func TestIfExpression(t *testing.T) {
	emitter := makeEmitter()
	emitter.emitExpression(&parser.IfExpression{
		Condition: &parser.Literal{
			Token: testToken{kind: parser.BooleanLiteral, value: "false"},
		},
		Body: &parser.Block{Statements: []parser.Node{}},
		Alternate: &parser.IfExpression{
			Condition: &parser.Literal{
				Token: testToken{kind: parser.BooleanLiteral, value: "false"},
			},
			Body:      &parser.Block{Statements: []parser.Node{}},
			Alternate: &parser.Block{},
		},
	})

	text := emitter.string()
	expected := "false ? undefined : false ? undefined : undefined"
	if text != expected {
		t.Fatalf("Expected string:\n%v\ngot:\n%v", expected, text)
	}
}

func TestEmitFunctionExpression(t *testing.T) {
	source := "triple :: (n number) => { 3 * n }"

	expected := "const triple = (n) => {\n"
	expected += "    return 3 * n;\n"
	expected += "}\n"

	ast, _ := parser.Parse(strings.NewReader(source))

	emitter := makeEmitter()
	emitter.emit(ast[0])
	received := emitter.string()
	if emitter.string() != expected {
		t.Fatalf("expected '%v', got '%v'", expected, received)
	}
}

func TestMapElementAccess(t *testing.T) {
	emitter := makeEmitter()
	emitMapElementAccess(emitter, &parser.ComputedAccessExpression{
		Expr: &parser.Identifier{Token: testToken{kind: parser.Name, value: "map"}},
		Property: &parser.BracketedExpression{
			Expr: &parser.Literal{Token: testToken{kind: parser.StringLiteral, value: "\"key\""}},
		},
	})

	text := emitter.string()
	expected := "map.get(\"key\")"
	if text != expected {
		t.Fatalf("Expected string:\n%v\ngot:\n%v", expected, text)
	}
}

func TestEmitReference(t *testing.T) {
	source := "value := 0\n"
	source += "&value"

	expected := "function (_) { return arguments.length ? void (value = _) : value }"

	ast, _ := parser.Parse(strings.NewReader(source))

	emitter := makeEmitter()
	emitter.emit(ast[1])
	received := emitter.string()
	if emitter.string() != expected {
		t.Fatalf("expected '%v', got '%v'", expected, received)
	}
}

func TestEmitArrayRef(t *testing.T) {
	source := "array := []number{0, 1, 2}\n"
	source += "&array\n"

	expected := "__slice(() => array)"

	ast, _ := parser.Parse(strings.NewReader(source))

	emitter := makeEmitter()
	emitter.emit(ast[1])
	received := emitter.string()
	if emitter.string() != expected {
		t.Fatalf("expected '%v', got '%v'", expected, received)
	}
}

func TestEmitSlice(t *testing.T) {
	source := "array := []number{0, 1, 2}\n"
	source += "&array[1..]\n"

	expected := "__slice(() => array, 1)"

	ast, _ := parser.Parse(strings.NewReader(source))

	emitter := makeEmitter()
	emitter.emit(ast[1])
	received := emitter.string()
	if emitter.string() != expected {
		t.Fatalf("expected '%v', got '%v'", expected, received)
	}
}

func TestEmitDeref(t *testing.T) {
	source := "value := 0\n"
	source += "ref := &value\n"
	source += "*ref"

	expected := "ref()"

	ast, _ := parser.Parse(strings.NewReader(source))

	emitter := makeEmitter()
	emitter.emit(ast[2])
	received := emitter.string()
	if emitter.string() != expected {
		t.Fatalf("expected '%v', got '%v'", expected, received)
	}
}
