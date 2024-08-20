package parser

import "github.com/bmelicque/test-parser/tokenizer"

type testTokenizer struct {
	tokens []tokenizer.Token
	index  int
}

type testToken struct {
	kind  tokenizer.TokenKind
	value string
	loc   tokenizer.Loc
}

func (t testToken) Kind() tokenizer.TokenKind { return t.kind }
func (t testToken) Text() string              { return t.value }
func (t testToken) Loc() tokenizer.Loc        { return t.loc }

func (t *testTokenizer) Dispose() {}
func (t *testTokenizer) Peek() tokenizer.Token {
	max := len(t.tokens)
	if t.index == max {
		loc := t.tokens[t.index-1].Loc()
		loc.Start = loc.End
		return testToken{tokenizer.EOF, "", loc}
	}
	return t.tokens[t.index]
}
func (t *testTokenizer) Consume() tokenizer.Token {
	max := len(t.tokens)
	if t.index == max {
		loc := t.tokens[t.index-1].Loc()
		loc.Start = loc.End
		return testToken{tokenizer.EOF, "", loc}
	}
	t.index++
	return t.tokens[t.index-1]
}
func (t *testTokenizer) DiscardLineBreaks() {
	for t.Peek().Kind() == tokenizer.EOL {
		t.Consume()
	}
}
