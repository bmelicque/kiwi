package emitter

import (
	"strings"
	"testing"

	"github.com/bmelicque/test-parser/parser"
)

type testToken struct {
	kind  parser.TokenKind
	value string
	loc   parser.Loc
}

func (t testToken) Kind() parser.TokenKind { return t.kind }
func (t testToken) Text() string           { return t.value }
func (t testToken) Loc() parser.Loc        { return t.loc }

func testEmitter(t *testing.T, source string, expected string, line int) {
	ast, err := parser.Parse(strings.NewReader(source))
	if len(err) > 0 {
		t.Log("Got unexpected parser errors:\n")
		for _, err := range err {
			line := err.Node.Loc().Start.Line
			col := err.Node.Loc().End.Col
			msg := err.Text()
			t.Logf("Error at line %v, col. %v: %v\n", line, col, msg)
		}
		t.FailNow()
	}
	emitter := makeEmitter()
	emitter.emit(ast[line])
	received := emitter.string()
	if emitter.string() != expected {
		t.Fatalf("expected output:\n%v\n\ngot:\n%v", expected, received)
	}
}
