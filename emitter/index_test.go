package emitter

import "github.com/bmelicque/test-parser/parser"

type testToken struct {
	kind  parser.TokenKind
	value string
	loc   parser.Loc
}

func (t testToken) Kind() parser.TokenKind { return t.kind }
func (t testToken) Text() string           { return t.value }
func (t testToken) Loc() parser.Loc        { return t.loc }
