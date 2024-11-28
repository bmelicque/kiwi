package emitter

import (
	"testing"

	"github.com/bmelicque/test-parser/parser"
)

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

func TestEmitSliceAccess(t *testing.T) {
	source := "array := []number{}\n"
	source += "slice := &array\n"
	source += "slice[0]"

	expected := "slice(0)"

	testEmitter(t, source, expected, 2)
}
