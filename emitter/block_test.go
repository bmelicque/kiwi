package emitter

import (
	"strings"
	"testing"

	"github.com/bmelicque/test-parser/parser"
)

func TestBlockSymbol(t *testing.T) {
	source := "{\n"
	source += "    value := 0\n"
	source += "    ref := &value\n"
	source += "}"
	ast, _, _ := parser.ParseProgram(strings.NewReader(source), "")
	block, ok := ast[0].(*parser.Block)
	if !ok {
		t.Fatalf("Block expected, got %#v", ast[0])
	}
	if !needsSymbol(block) {
		t.Fatalf("Expected block to need a symbol")
	}
}

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
