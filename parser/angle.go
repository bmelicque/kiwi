package parser

import (
	"github.com/bmelicque/test-parser/tokenizer"
)

// Expression between angle brackets
type AngleExpression struct {
	Expr Node
	loc  tokenizer.Loc
}

func (a AngleExpression) Loc() tokenizer.Loc { return a.loc }
func (p *Parser) parseAngleExpression() AngleExpression {
	loc := p.tokenizer.Consume().Loc()

	outer := p.allowAngleBrackets
	p.allowAngleBrackets = false
	expr := ParseExpression(p)
	p.allowAngleBrackets = outer

	next := p.tokenizer.Peek()
	if next.Kind() != tokenizer.GREATER {
		p.report("Expected '>'", next.Loc())
		if expr != nil {
			loc.End = expr.Loc().End
		}
	} else {
		loc.End = p.tokenizer.Consume().Loc().End
	}

	return AngleExpression{expr, loc}
}
