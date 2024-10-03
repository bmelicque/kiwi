package emitter

import (
	"testing"

	"github.com/bmelicque/test-parser/checker"
	"github.com/bmelicque/test-parser/parser"
	"github.com/bmelicque/test-parser/tokenizer"
)

func TestIfElse(t *testing.T) {
	emitter := makeEmitter()
	emitter.emit(checker.If{
		Condition: checker.Literal{parser.TokenExpression{Token: testToken{kind: tokenizer.BOOLEAN, value: "false"}}},
		Body:      checker.Body{Statements: []checker.Node{}},
		Alternate: checker.If{
			Condition: checker.Literal{parser.TokenExpression{Token: testToken{kind: tokenizer.BOOLEAN, value: "false"}}},
			Body:      checker.Body{Statements: []checker.Node{}},
			Alternate: checker.Body{},
		},
	})

	text := emitter.string()
	expected := "if (false) {} else if (false) {} else {}"
	if text != expected {
		t.Fatalf("Expected string:\n%v\ngot:\n%v", expected, text)
	}
}
