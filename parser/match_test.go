package parser

import (
	"strings"
	"testing"
)

func TestMatch(t *testing.T) {
	str := "match option {\n"
	str += "case Some(s):\n"
	str += "    return s\n"
	str += "}"
	parser, err := MakeParser(strings.NewReader(str))
	if err != nil {
		t.Fatal(err)
	}
	node := parser.parseMatchExpression()
	statement, ok := node.(*MatchExpression)
	if !ok {
		t.Fatalf("Expected match expression, got %#v", node)
	}

	if len(statement.Cases) != 1 {
		t.Fatalf("Expected 1 case, got %#v", statement.Cases)
	}
}
