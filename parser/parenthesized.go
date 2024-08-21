package parser

import (
	"github.com/bmelicque/test-parser/tokenizer"
)

type ParenthesizedExpression struct {
	Expr Node
	loc  tokenizer.Loc
}

func (p ParenthesizedExpression) Loc() tokenizer.Loc {
	return p.loc
}

func (p *Parser) parseParenthesizedExpression() ParenthesizedExpression {
	loc := p.tokenizer.Consume().Loc() // LPAREN
	next := p.tokenizer.Peek()
	if next.Kind() == tokenizer.RPAREN {
		loc.End = p.tokenizer.Consume().Loc().End
		return ParenthesizedExpression{nil, loc}
	}
	outer := p.allowBraceParsing
	p.allowBraceParsing = true
	expr := ParseExpression(p)
	p.allowBraceParsing = outer
	next = p.tokenizer.Peek()
	if next.Kind() != tokenizer.RPAREN {
		p.report("')' expected", next.Loc())
	}
	loc.End = p.tokenizer.Consume().Loc().End
	return ParenthesizedExpression{expr, loc}
}
