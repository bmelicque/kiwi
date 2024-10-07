package emitter

import (
	"testing"

	"github.com/bmelicque/test-parser/checker"
	"github.com/bmelicque/test-parser/parser"
	"github.com/bmelicque/test-parser/tokenizer"
)

func TestIfExpression(t *testing.T) {
	emitter := makeEmitter()
	emitter.emit(checker.If{
		Condition: checker.Literal{TokenExpression: parser.TokenExpression{
			Token: testToken{kind: tokenizer.BOOLEAN, value: "false"},
		}},
		Block: checker.Block{Statements: []checker.Node{}},
		Alternate: checker.If{
			Condition: checker.Literal{TokenExpression: parser.TokenExpression{
				Token: testToken{kind: tokenizer.BOOLEAN, value: "false"},
			}},
			Block:     checker.Block{Statements: []checker.Node{}},
			Alternate: checker.Block{},
		},
	})

	text := emitter.string()
	expected := "false ? undefined : false ? undefined : undefined"
	if text != expected {
		t.Fatalf("Expected string:\n%v\ngot:\n%v", expected, text)
	}
}
