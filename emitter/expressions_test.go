package emitter

import (
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
