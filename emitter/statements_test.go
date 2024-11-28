package emitter

import (
	"testing"

	"github.com/bmelicque/test-parser/parser"
)

func TestIfStatement(t *testing.T) {
	emitter := makeEmitter()
	emitter.emit(&parser.IfExpression{
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
	expected := "if (false) {} else if (false) {} else {}"
	if text != expected {
		t.Fatalf("Expected string:\n%v\ngot:\n%v", expected, text)
	}
}
