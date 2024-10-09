package parser

type TypedExpression struct {
	Expr   Expression
	Typing Expression
	Colon  bool
}

func (t TypedExpression) Loc() Loc {
	loc := t.Expr.Loc()
	if t.Typing != nil {
		loc.End = t.Typing.Loc().End
	}
	return loc
}

// FIXME:
func (t TypedExpression) Type() ExpressionType { return nil }

func (p *Parser) parseTypedExpression() Expression {
	expr := p.parseRange()
	colon := false
	if p.Peek().Kind() == Colon {
		p.Consume()
		colon = true
	}
	outer := p.allowEmptyExpr
	if !colon {
		p.allowEmptyExpr = true
	}
	typing := p.parseRange()
	p.allowEmptyExpr = outer
	if typing == nil {
		return expr
	}
	return TypedExpression{expr, typing, colon}
}
