package parser

import (
	"strings"
	"testing"
)

func TestMapIdentifier(t *testing.T) {
	parser, err := MakeParser(strings.NewReader(""))
	if err != nil {
		t.Fatal(err)
	}
	expr := &Identifier{Token: literal{kind: Name, value: "Map"}}
	expr.typeCheck(parser)

	if _, ok := expr.Type().(Type); !ok {
		t.Fatalf("Type expected")
	}
}
