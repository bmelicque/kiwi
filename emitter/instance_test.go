package emitter

import (
	"strings"
	"testing"

	"github.com/bmelicque/test-parser/parser"
)

func TestEmitInstanceExpression(t *testing.T) {
	type test struct {
		name     string
		src      string
		expected string
	}
	tests := []test{
		{
			name:     "option with no arg",
			src:      "?number{}",
			expected: "new __.Option(\"None\");\n",
		},
		{
			name:     "option with arg",
			src:      "?number{42}",
			expected: "new __.Option(\"Some\", 42);\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			program, _ := parser.ParseProgram(strings.NewReader(tt.src), "")

			emitter := makeEmitter()
			emitter.emit(program.Nodes()[0])
			text := emitter.string()

			if text != tt.expected {
				t.Errorf("Expected statement:\n%v\ngot:\n%v", tt.expected, text)
			}
		})
	}

}

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
