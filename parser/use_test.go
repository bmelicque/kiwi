package parser

import (
	"strings"
	"testing"
)

func TestParseUseIdentifier(t *testing.T) {
	str := "use log from \"io\""
	parser := MakeParser(strings.NewReader(str))
	parser.parseAssignment()
	testParserErrors(t, parser, 0)
}

func TestParseUseTuple(t *testing.T) {
	str := "use log, debug from \"io\""
	parser := MakeParser(strings.NewReader(str))
	parser.parseAssignment()
	testParserErrors(t, parser, 0)
}

func TestParseUseAs(t *testing.T) {
	str := "use * as io from \"io\""
	parser := MakeParser(strings.NewReader(str))
	parser.parseAssignment()
	testParserErrors(t, parser, 0)
}
