package emitter

import "testing"

func TestEmitFor(t *testing.T) {
	source := "for {}"
	expected := "while (true) {}"
	testEmitter(t, source, expected, 0)
}

func TestEmitForCondition(t *testing.T) {
	source := "for true {}"
	expected := "while (true) {}"
	testEmitter(t, source, expected, 0)
}

func TestEmitForInRange(t *testing.T) {
	source := "for x in 3..10 { x }"
	expected := "for (let x = 3; x < 10; x++) {\n"
	expected += "    x;\n"
	expected += "}\n"
	testEmitter(t, source, expected, 0)
}

func TestEmitForInRangeTuple(t *testing.T) {
	source := "for x, i in 3..=10 { x + i }"
	expected := "for (let x = 3, i = 0; x <= 10; x++, i++) {\n"
	expected += "    x + i;\n"
	expected += "}\n"
	testEmitter(t, source, expected, 0)
}

func TestEmitForInList(t *testing.T) {
	source := "list := []number{1, 2, 3}\n"
	source += "for x in list { x }"

	expected := "for (let x of list) {\n"
	expected += "    x;\n"
	expected += "}\n"
	testEmitter(t, source, expected, 1)
}

func TestEmitForInListTuple(t *testing.T) {
	source := "list := []number{1, 2, 3}\n"
	source += "for x, i in list { x + i }"

	expected := "const __list = list;\n"
	expected += "for (let x = __list[0], i = 0; i < __list.length; x = __list[++i]) {\n"
	expected += "    x + i;\n"
	expected += "}\n"
	testEmitter(t, source, expected, 1)
}

func TestEmitForInSlice(t *testing.T) {
	source := "list := []number{1, 2, 3}\n"
	source += "slice := &list\n"
	source += "for x in slice { x }"

	expected := "for (let x of slice) {\n"
	expected += "    x;\n"
	expected += "}\n"
	testEmitter(t, source, expected, 2)
}

func TestEmitForInSliceTuple(t *testing.T) {
	source := "list := []number{1, 2, 3}\n"
	source += "slice := &list\n"
	source += "for x, i in slice { *x + i }"

	expected := "const __s = slice;\n"
	expected += "for (let x = __s.ref(0), i = 0; i < __s.length; x = __s.ref(++i)) {\n"
	expected += "    x() + i;\n"
	expected += "}\n"
	testEmitter(t, source, expected, 2)
}
