package emitter

import "testing"

func TestEmitBinaryExpression(t *testing.T) {
	source := "1 + 2"
	expected := "1 + 2"
	testEmitter(t, source, expected, 0)
}
