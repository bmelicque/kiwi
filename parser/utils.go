package parser

import (
	"fmt"
	"slices"
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
