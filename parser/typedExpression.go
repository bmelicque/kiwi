package parser

import (
	"github.com/bmelicque/test-parser/tokenizer"
)

type TypedExpression struct {
	Expr     Node
	operator tokenizer.Token
	Typing   Node
}

func (t TypedExpression) Loc() tokenizer.Loc {
	loc := t.operator.Loc()
	if t.Expr != nil {
		loc.Start = t.Expr.Loc().Start
	}
	if t.Typing != nil {
		loc.End = t.Typing.Loc().End
	}
	return loc
}

func ParseTypedExpression(p *Parser) Node {
	expr := ParseExpression(p)
	if p.tokenizer.Peek().Kind() != tokenizer.COLON {
		return expr
	}
	operator := p.tokenizer.Consume()
	typing := ParseExpression(p)
	return TypedExpression{expr, operator, typing}
}
