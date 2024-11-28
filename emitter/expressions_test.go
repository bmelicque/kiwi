package emitter

import (
	"testing"

	"github.com/bmelicque/test-parser/parser"
)

func TestEmitFunctionExpression(t *testing.T) {
	source := "triple :: (n number) => { 3 * n }"

	expected := "const triple = (n) => {\n"
	expected += "    return 3 * n;\n"
	expected += "}\n"

	testEmitter(t, source, expected, 0)
}

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

func TestEmitReference(t *testing.T) {
	source := "value := 0\n"
	source += "&value"

	expected := "function (_) { return arguments.length ? void (value = _) : value }"

	testEmitter(t, source, expected, 1)
}

func TestEmitArrayRef(t *testing.T) {
	source := "array := []number{0, 1, 2}\n"
	source += "&array\n"

	expected := "__slice(() => array)"

	testEmitter(t, source, expected, 1)
}

func TestEmitSlice(t *testing.T) {
	source := "array := []number{0, 1, 2}\n"
	source += "&array[1..]\n"

	expected := "__slice(() => array, 1)"

	testEmitter(t, source, expected, 1)
}

func TestEmitDeref(t *testing.T) {
	source := "value := 0\n"
	source += "ref := &value\n"
	source += "*ref"

	expected := "ref()"

	testEmitter(t, source, expected, 2)
}
