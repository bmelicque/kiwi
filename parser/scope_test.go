package parser

import "testing"

func TestMapIdentifier(t *testing.T) {
	parser := MakeParser(nil)
	expr := &Identifier{Token: literal{kind: Name, value: "Map"}}
	expr.typeCheck(parser)

	if expr.Type().Kind() != TYPE {
		t.Fatalf("Type expected")
	}
}
