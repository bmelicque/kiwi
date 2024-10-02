package parser

import (
	"github.com/bmelicque/test-parser/tokenizer"
)

type TypedExpression struct {
	Expr   Node
	Typing Node
	Colon  bool
}

func (t TypedExpression) Loc() tokenizer.Loc {
	loc := t.Expr.Loc()
	if t.Typing != nil {
		loc.End = t.Typing.Loc().End
	}
	return loc
}

func (p *Parser) parseTypedExpression() Node {
	expr := ParseRange(p)
	colon := false
	if p.tokenizer.Peek().Kind() == tokenizer.COLON {
		p.tokenizer.Consume()
		colon = true
	}
	outer := p.allowEmptyExpr
	if !colon {
		p.allowEmptyExpr = true
	}
	typing := ParseRange(p)
	p.allowEmptyExpr = outer
	if typing == nil {
		return expr
	}
	return TypedExpression{expr, typing, colon}
}
