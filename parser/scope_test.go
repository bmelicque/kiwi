package parser

import (
	"testing"
)

func TestMapIdentifier(t *testing.T) {
	parser, _ := MakeParser(nil)
	expr := &Identifier{Token: literal{kind: Name, value: "Map"}}
	expr.typeCheck(parser)

	if _, ok := expr.Type().(Type); !ok {
		t.Fatalf("Type expected")
	}
}
