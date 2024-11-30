package emitter

import (
	"testing"
)

func TestEmitReference(t *testing.T) {
	source := "value := 0\n"
	source += "&value"

	expected := "function (_) { return arguments.length ? void (value = _) : value }"

	testEmitter(t, source, expected, 1)
}

func TestEmitListRef(t *testing.T) {
	source := "array := []number{0, 1, 2}\n"
	source += "&array\n"

	expected := "new __Slice(() => array)"

	testEmitter(t, source, expected, 1)
}

func TestEmitSlice(t *testing.T) {
	source := "array := []number{0, 1, 2}\n"
	source += "&array[1..]\n"

	expected := "new __Slice(() => array, 1)"

	testEmitter(t, source, expected, 1)
}

func TestEmitDeref(t *testing.T) {
	source := "value := 0\n"
	source += "ref := &value\n"
	source += "*ref"

	expected := "ref()"

	testEmitter(t, source, expected, 2)
}
