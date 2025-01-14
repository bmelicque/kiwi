package emitter

import (
	"testing"
)

func TestAssignmentShorthand(t *testing.T) {
	source := "_n := 42\n"
	source += "_n += 1\n"

	expected := "_n += 1;\n"

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
	source := "_BoxedNumber :: { value number }"
	expected := "class _BoxedNumber {\n"
	expected += "    constructor(value) {\n"
	expected += "        this.value = value;\n"
	expected += "    }\n"
	expected += "}\n"
	testEmitter(t, source, expected, 0)
}

func TestObjectDefinitionDefault(t *testing.T) {
	source := "_BoxedNumber :: { value: 0 }"
	expected := "class _BoxedNumber {\n"
	expected += "    constructor(value = 0) {\n"
	expected += "        this.value = value;\n"
	expected += "    }\n"
	expected += "}\n"
	testEmitter(t, source, expected, 0)
}

func TestGenericObjectDefintion(t *testing.T) {
	source := "_Boxed[Type] :: { value Type }"
	expected := "class _Boxed {\n"
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
