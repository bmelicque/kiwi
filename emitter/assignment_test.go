package emitter

import (
	"testing"
)

func TestAssignmentShorthand(t *testing.T) {
	source := "n := 42\n"
	source += "n += 1\n"

	expected := "n += 1;\n"

	testEmitter(t, source, expected, 1)
}

func TestInderectAssignment(t *testing.T) {
	source := "i := 0\n"
	source += "ref := &i\n"
	source += "*ref = 42"

	expected := "ref(0, 42)"

	testEmitter(t, source, expected, 2)
}

func TestObjectDefinition(t *testing.T) {
	source := "BoxedNumber :: { value number }"
	expected := "class BoxedNumber {\n"
	expected += "    constructor(value) {\n"
	expected += "        this.value = value;\n"
	expected += "    }\n"
	expected += "}\n"
	testEmitter(t, source, expected, 0)
}

func TestObjectDefinitionDefault(t *testing.T) {
	source := "BoxedNumber :: { value: 0 }"
	expected := "class BoxedNumber {\n"
	expected += "    constructor(value = 0) {\n"
	expected += "        this.value = value;\n"
	expected += "    }\n"
	expected += "}\n"
	testEmitter(t, source, expected, 0)
}

func TestGenericObjectDefintion(t *testing.T) {
	source := "Boxed[Type] :: { value Type }"
	expected := "class Boxed {\n"
	expected += "    constructor(value) {\n"
	expected += "        this.value = value;\n"
	expected += "    }\n"
	expected += "}\n"
	testEmitter(t, source, expected, 0)
}

func TestMethodDefinition(t *testing.T) {
	source := "User :: { name string }\n"
	source += "(u User).getName :: () => { u.name }"
	expected := "User.prototype.getName = function () {\n"
	expected += "    return this.name;\n"
	expected += "}\n"
	testEmitter(t, source, expected, 1)
}
