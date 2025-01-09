package emitter

import (
	"testing"

	"github.com/bmelicque/test-parser/parser"
)

func TestCatchStatement(t *testing.T) {
	emitter := makeEmitter()
	// result catch _ { 0 }
	emitter.emit(&parser.CatchExpression{
		Left:       &parser.Identifier{Token: testToken{kind: parser.Name, value: "result"}},
		Identifier: &parser.Identifier{Token: testToken{kind: parser.Name, value: "_"}},
		Body: parser.MakeBlock([]parser.Node{
			&parser.Literal{Token: testToken{kind: parser.NumberLiteral, value: "0"}},
		}),
	})

	text := emitter.string()
	expected := "try {\n"
	expected += "    result;\n"
	expected += "} catch (_) {\n"
	expected += "    0;\n"
	expected += "}\n"
	if text != expected {
		t.Fatalf("Expected string:\n%v\ngot:\n%v", expected, text)
	}
}
