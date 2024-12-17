package parser

import (
	"strings"
	"testing"
)

func TestMapIdentifier(t *testing.T) {
	parser := MakeParser(nil)
	expr := &Identifier{Token: literal{kind: Name, value: "Map"}}
	expr.typeCheck(parser)

	if _, ok := expr.Type().(Type); !ok {
		t.Fatalf("Type expected")
	}
}

func TestStdIO(t *testing.T) {
	source := "io.log(42)"
	parser := MakeParser(strings.NewReader(source))
	parser.parseExpression().typeCheck(parser)
	testParserErrors(t, parser, 0)
}
