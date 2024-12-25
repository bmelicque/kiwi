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

func TestGenericObjectImplicitInstance(t *testing.T) {
	source := "Boxed[Type] :: {\n"
	source += "    value Type\n"
	source += "}\n"
	source += "Boxed{ value: 42 }"

	expected := "new Boxed(42);\n"

	testEmitter(t, source, expected, 1)
}

func TestGenericObjectExplicitInstance(t *testing.T) {
	source := "Boxed[Type] :: {\n"
	source += "    value Type\n"
	source += "}\n"
	source += "Boxed[number]{ value: 42 }"

	expected := "new Boxed(42);\n"

	testEmitter(t, source, expected, 1)
}

func TestMapInstance(t *testing.T) {
	source := "Map{ \"value\": 42 }"
	expected := "new Map([[\"value\", 42]]);\n"
	testEmitter(t, source, expected, 0)
}
