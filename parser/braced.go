package parser

// Expression between braces, such as `{value}`
type BracedExpression struct {
	Expr Expression
	loc  Loc
}

func (b *BracedExpression) getChildren() []Node {
	if b.Expr == nil {
		return []Node{}
	}
	return []Node{b.Expr}
}

func (b *BracedExpression) typeCheck(p *Parser) {
	b.Expr.typeCheck(p)
}

func (b *BracedExpression) Loc() Loc             { return b.loc }
func (b *BracedExpression) Type() ExpressionType { return nil }
func (p *Parser) parseBracedExpression() *BracedExpression {
	if p.Peek().Kind() != LeftBrace {
		panic("'{' expected!")
	}
	loc := p.Consume().Loc()

	outer := p.allowEmptyExpr
	p.allowEmptyExpr = true
	expr := p.parseExpression()
	p.allowEmptyExpr = outer

	next := p.Peek()
	if next.Kind() != RightBrace {
		p.report("'}' expected", next.Loc())
		if expr != nil {
			loc.End = expr.Loc().End
		}
	} else {
		loc.End = p.Consume().Loc().End
	}

	return &BracedExpression{expr, loc}
}
