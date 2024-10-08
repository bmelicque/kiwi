package parser

import (
	"fmt"
	"slices"
	"unicode"
)

func recover(p *Parser, at TokenKind) bool {
	next := p.Peek()
	start := next.Loc().Start
	end := start
	recovery := []TokenKind{at, EOL, EOF}
	for ; slices.Contains(recovery, next.Kind()); next = p.Peek() {
		end = p.Consume().Loc().End
	}
	// FIXME: token text
	p.report(fmt.Sprintf("'%v' expected", at), Loc{Start: start, End: end})
	return next.Kind() == LeftBrace
}

func IsTypeToken(expr Node) bool {
	token, ok := expr.(TokenExpression)
	if !ok {
		return false
	}

	if token.Token.Kind() != Name {
		return false
	}

	return unicode.IsUpper(rune(token.Token.Text()[0]))
}
