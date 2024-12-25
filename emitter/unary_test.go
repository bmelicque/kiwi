package emitter

import (
	"testing"
)

func TestEmitReference(t *testing.T) {
	source := "value := 0\n"
	source += "&value"

	expected := "(_,__)=>(_&4?__s:_&2?\"value\":_?value:(value=__));\n"

	testEmitter(t, source, expected, 1)
}

func TestEmitDeref(t *testing.T) {
	source := "value := 0\n"
	source += "ref := &value\n"
	source += "*ref"

	expected := "ref(1);\n"

	testEmitter(t, source, expected, 2)
}
