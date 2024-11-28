package emitter

import (
	"testing"

	"github.com/bmelicque/test-parser/parser"
)

func TestMapAssignment(t *testing.T) {
	emitter := makeEmitter()
	// map[key] = value
	emitSetMap(emitter, &parser.Assignment{
		Pattern: &parser.ComputedAccessExpression{
			Expr: &parser.Identifier{Token: testToken{kind: parser.Name, value: "map"}},
			Property: &parser.BracketedExpression{
				Expr: &parser.Literal{Token: testToken{kind: parser.StringLiteral, value: "\"key\""}},
			},
		},
		Value:    &parser.Literal{Token: testToken{kind: parser.NumberLiteral, value: "42"}},
		Operator: testToken{kind: parser.Assign},
	})

	text := emitter.string()
	expected := "map.set(\"key\", 42)"
	if text != expected {
		t.Fatalf("Expected string:\n%v\ngot:\n%v", expected, text)
	}
}

func TestSliceAssignment(t *testing.T) {
	source := "array := []number{}\n"
	source += "slice := &array\n"
	source += "slice[0] = 42"

	expected := "slice(0, 42)"

	testEmitter(t, source, expected, 2)
}
