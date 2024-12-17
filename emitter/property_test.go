package emitter

import "testing"

func TestEmitPropertyAccess(t *testing.T) {
	source := "Type :: { value number }\n"
	source += "x := Type{ value: 42 }\n"
	source += "x.value"

	expected := "x.value;\n"

	testEmitter(t, source, expected, 2)
}

func TestEmitIndirectPropertyAccess(t *testing.T) {
	source := "Type :: { value number }\n"
	source += "x := Type{ value: 42 }\n"
	source += "ref := &x\n"
	source += "ref.value"

	expected := "ref(1).value;\n"

	testEmitter(t, source, expected, 3)
}
