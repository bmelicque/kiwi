package parser

import (
	"github.com/bmelicque/test-parser/tokenizer"
)

type TypedExpression struct {
	Expr   Node
	Typing Node
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
	if p.tokenizer.Peek().Kind() == tokenizer.COLON {
		p.report("Params and members don't use colons", p.tokenizer.Consume().Loc())
	}
	p.allowEmptyExpr = true
	typing := ParseRange(p)
	p.allowEmptyExpr = false
	if typing == nil {
		return expr
	}
	return TypedExpression{expr, typing}
}
