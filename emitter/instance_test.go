package emitter

import (
	"testing"
)

func TestObjectInstance(t *testing.T) {
	source := "Boxed :: {\n"
	source += "    value number\n"
	source += "}\n"
	source += "Boxed{ value: 42 }"

	expected := "new Boxed(42);\n"

	testEmitter(t, source, expected, 1)
}

func TestObjectInstanceWithOptionals(t *testing.T) {
	source := "Boxed :: {\n"
	source += "    value    number\n"
	source += "    default: 42\n"
	source += "}\n"
	source += "Boxed{ value: 42 }"

	expected := "new Boxed(42);\n"

	testEmitter(t, source, expected, 1)
}
