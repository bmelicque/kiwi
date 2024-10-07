package emitter

import (
	"testing"

	"github.com/bmelicque/test-parser/checker"
	"github.com/bmelicque/test-parser/parser"
	"github.com/bmelicque/test-parser/tokenizer"
)

func TestEmptyBlockExpression(t *testing.T) {
	emitter := makeEmitter()
	emitter.emit(checker.Block{})

	text := emitter.string()
	expected := "undefined"
	if text != expected {
		t.Fatalf("Expected string:\n%v\ngot:\n%v", expected, text)
	}
}

func TestSingleLineBlockExpression(t *testing.T) {
	emitter := makeEmitter()
	emitter.emit(checker.Block{Statements: []checker.Node{
		checker.Literal{TokenExpression: parser.TokenExpression{
			Token: testToken{kind: tokenizer.NUMBER, value: "42"},
		}},
	}})

	text := emitter.string()
	expected := "42"
	if text != expected {
		t.Fatalf("Expected string:\n%v\ngot:\n%v", expected, text)
	}
}

func TestBlockExpression(t *testing.T) {
	emitter := makeEmitter()
	emitter.emit(checker.Block{Statements: []checker.Node{
		checker.Literal{TokenExpression: parser.TokenExpression{
			Token: testToken{kind: tokenizer.NUMBER, value: "42"},
		}},
		checker.Literal{TokenExpression: parser.TokenExpression{
			Token: testToken{kind: tokenizer.NUMBER, value: "42"},
		}},
	}})

	text := emitter.string()
	expected := "(\n    42,\n    42,\n)"
	if text != expected {
		t.Fatalf("Expected string:\n%v\ngot:\n%v", expected, text)
	}
}

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
