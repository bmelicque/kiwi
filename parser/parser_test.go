package parser

type testTokenizer struct {
	tokens []Token
	index  int
}

func (t *testTokenizer) Dispose() {}
func (t *testTokenizer) Peek() Token {
	max := len(t.tokens)
	if t.index == max {
		loc := t.tokens[t.index-1].Loc()
		loc.Start = loc.End
		return token{EOF, loc}
	}
	return t.tokens[t.index]
}
func (t *testTokenizer) Consume() Token {
	max := len(t.tokens)
	if t.index == max {
		loc := t.tokens[t.index-1].Loc()
		loc.Start = loc.End
		return token{EOF, loc}
	}
	t.index++
	return t.tokens[t.index-1]
}
func (t *testTokenizer) DiscardLineBreaks() {
	for t.Peek().Kind() == EOL {
		t.Consume()
	}
}
