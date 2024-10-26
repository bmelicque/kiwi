package emitter

import (
	"testing"

	"github.com/bmelicque/test-parser/parser"
)

func TestExtractBlock(t *testing.T) {
	emitter := makeEmitter()
	emitter.extractUninlinables(&parser.Block{Statements: []parser.Node{
		&parser.Literal{Token: testToken{kind: parser.NumberLiteral, value: "0"}},
		&parser.Literal{Token: testToken{kind: parser.NumberLiteral, value: "1"}},
		&parser.Literal{Token: testToken{kind: parser.NumberLiteral, value: "2"}},
	}})

	text := emitter.string()
	expected := "let _tmp0;\n"
	expected += "{\n"
	expected += "    0;\n"
	expected += "    1;\n"
	expected += "    _tmp0 = 2;\n"
	expected += "}\n"
	if text != expected {
		t.Fatalf("Expected string:\n%v\ngot:\n%v", expected, text)
	}
}

func TestExtractCatch(t *testing.T) {
	emitter := makeEmitter()
	// result catch _ { 0 }
	emitter.extractUninlinables(&parser.CatchExpression{
		Left:       &parser.Identifier{Token: testToken{kind: parser.Name, value: "result"}},
		Identifier: &parser.Identifier{Token: testToken{kind: parser.Name, value: "_"}},
		Body: &parser.Block{Statements: []parser.Node{
			&parser.Literal{Token: testToken{kind: parser.NumberLiteral, value: "0"}},
		}},
	})

	text := emitter.string()
	expected := "let _tmp0;\n"
	expected += "try {\n"
	expected += "    _tmp0 = result;\n"
	expected += "} catch (_) {\n"
	expected += "    _tmp0 = 0;\n"
	expected += "}\n"
	if text != expected {
		t.Fatalf("Expected string:\n%v\ngot:\n%v", expected, text)
	}
}
