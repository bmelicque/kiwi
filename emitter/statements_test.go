package emitter

import (
	"testing"

	"github.com/bmelicque/test-parser/checker"
	"github.com/bmelicque/test-parser/parser"
)

func TestIfStatement(t *testing.T) {
	emitter := makeEmitter()
	emitter.emit(checker.ExpressionStatement{Expr: checker.If{
		Condition: checker.Literal{TokenExpression: parser.TokenExpression{
			Token: testToken{kind: parser.BOOLEAN, value: "false"},
		}},
		Block: checker.Block{Statements: []checker.Node{}},
		Alternate: checker.If{
			Condition: checker.Literal{TokenExpression: parser.TokenExpression{
				Token: testToken{kind: parser.BOOLEAN, value: "false"},
			}},
			Block:     checker.Block{Statements: []checker.Node{}},
			Alternate: checker.Block{},
		},
	}})

	text := emitter.string()
	expected := "if (false) {} else if (false) {} else {}"
	if text != expected {
		t.Fatalf("Expected string:\n%v\ngot:\n%v", expected, text)
	}
}
