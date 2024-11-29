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
