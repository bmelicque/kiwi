package emitter

import (
	"testing"
)

func TestEmitReference(t *testing.T) {
	source := "value := 0\n"
	source += "&value"

	expected := "(a,p)=>(a&4?__s:a&2?\"value\":a?value:(value=p));\n"

	testEmitter(t, source, expected, 1)
}

func TestEmitDeref(t *testing.T) {
	source := "value := 0\n"
	source += "ref := &value\n"
	source += "*ref"

	expected := "ref(1);\n"

	testEmitter(t, source, expected, 2)
}
