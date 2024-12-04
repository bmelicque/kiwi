package emitter

import (
	"testing"

	"github.com/bmelicque/test-parser/parser"
)

func TestMapAssignment(t *testing.T) {
	emitter := makeEmitter()
	// map[key] = value
	emitSetElement(emitter, &parser.Assignment{
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

	expected := "slice.set(0, 42)"

	testEmitter(t, source, expected, 2)
}

func TestSliceDerefAssignment(t *testing.T) {
	source := "array := []number{}\n"
	source += "slice := &array\n"
	source += "array = *slice"

	expected := "array = slice.clone();"

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
