package parser

// Expression between brackets, such as `[Type]`
type BracketedExpression struct {
	Expr Node
	loc  Loc
}

func (b BracketedExpression) Loc() Loc { return b.loc }
func (p *Parser) parseBracketedExpression() BracketedExpression {
	if p.Peek().Kind() != LBRACKET {
		panic("'[' expected!")
	}
	loc := p.Consume().Loc()

	outer := p.allowEmptyExpr
	p.allowEmptyExpr = true
	expr := ParseExpression(p)
	p.allowEmptyExpr = outer

	next := p.Peek()
	if next.Kind() != RBRACKET {
		p.report("']' expected", next.Loc())
		if expr != nil {
			loc.End = expr.Loc().End
		}
	} else {
		loc.End = p.Consume().Loc().End
	}

	return BracketedExpression{expr, loc}
}
