package emitter

import (
	"testing"
)

func TestEmitReference(t *testing.T) {
	source := "value := 0\n"
	source += "&value"

	// FIXME: __s55 is magic, should handle this properly (regex?)
	expected := "new __.Pointer(__s54, \"value\");\n"

	testEmitter(t, source, expected, 1)
}

func TestEmitDeref(t *testing.T) {
	source := "value := 0\n"
	source += "ref := &value\n"
	source += "*ref"

	expected := "ref(1);\n"

	testEmitter(t, source, expected, 2)
}
