package emitter

import (
	"regexp"
	"strings"
	"testing"

	"github.com/bmelicque/test-parser/parser"
)

func TestEmitReference(t *testing.T) {
	source := "value := 0\n"
	source += "&value"

	program, _ := parser.ParseProgram(strings.NewReader(source), "")
	emitter := makeEmitter()
	emitter.emit(program.Nodes()[1])
	received := emitter.string()
	matched, _ := regexp.Match(`new __\.Pointer\(__s\d+, "value"\);`, []byte(received))
	if !matched {
		t.Fatalf(
			"expected output:\n%v\n\ngot:\n%v",
			"new __.Pointer(__sXX, \"value\");\n",
			received,
		)
	}
}

func TestEmitDeref(t *testing.T) {
	source := "value := 0\n"
	source += "ref := &value\n"
	source += "*ref"

	expected := "ref(1);\n"

	testEmitter(t, source, expected, 2)
}
