package emitter

import "testing"

func TestEmitBinaryExpression(t *testing.T) {
	source := "1 + 2"
	expected := "1 + 2;\n"
	testEmitter(t, source, expected, 0)
}

func TestEmitRefComparison(t *testing.T) {
	source := "value := 0\n"
	source += "a := &value\n"
	source += "b := &value\n"
	source += "a == b"

	expected := "__.equals(a, b);\n"
	testEmitter(t, source, expected, 3)
}
