package emitter

import "testing"

func TestEmitFunctionExpression(t *testing.T) {
	source := "triple :: (n number) => { 3 * n }"

	expected := "const triple = (n) => {\n"
	expected += "    return 3 * n;\n"
	expected += "}\n"

	testEmitter(t, source, expected, 0)
}