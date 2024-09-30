package parser

import (
	"fmt"
	"slices"
	"unicode"

	"github.com/bmelicque/test-parser/tokenizer"
)

func recover(p *Parser, at tokenizer.TokenKind) bool {
	next := p.tokenizer.Peek()
	start := next.Loc().Start
	end := start
	recovery := []tokenizer.TokenKind{at, tokenizer.EOL, tokenizer.EOF}
	for ; slices.Contains(recovery, next.Kind()); next = p.tokenizer.Peek() {
		end = p.tokenizer.Consume().Loc().End
	}
	// FIXME: token text
	p.report(fmt.Sprintf("'%v' expected", at), tokenizer.Loc{Start: start, End: end})
	return next.Kind() == tokenizer.LBRACE
}

func IsTypeToken(expr Node) bool {
	token, ok := expr.(TokenExpression)
	if !ok {
		return false
	}

	if token.Token.Kind() != tokenizer.IDENTIFIER {
		return false
	}

	return unicode.IsUpper(rune(token.Token.Text()[0]))
}
