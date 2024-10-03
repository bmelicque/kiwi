package emitter

import "github.com/bmelicque/test-parser/tokenizer"

type testToken struct {
	kind  tokenizer.TokenKind
	value string
	loc   tokenizer.Loc
}

func (t testToken) Kind() tokenizer.TokenKind { return t.kind }
func (t testToken) Text() string              { return t.value }
func (t testToken) Loc() tokenizer.Loc        { return t.loc }
