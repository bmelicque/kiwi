package parser

// Expression between brackets, such as `[Type]`
type BracketedExpression struct {
	Expr Expression
	loc  Loc
}

func (b *BracketedExpression) getChildren() []Node {
	if b.Expr == nil {
		return []Node{}
	}
	return []Node{b.Expr}
}

func (b *BracketedExpression) typeCheck(p *Parser) {
	b.Expr.typeCheck(p)
}

func (b *BracketedExpression) Loc() Loc             { return b.loc }
func (b *BracketedExpression) Type() ExpressionType { return nil }
func (p *Parser) parseBracketedExpression() *BracketedExpression {
	if p.Peek().Kind() != LeftBracket {
		panic("'[' expected!")
	}
	loc := p.Consume().Loc()

	outer := p.allowEmptyExpr
	p.allowEmptyExpr = true
	expr := ParseExpression(p)
	p.allowEmptyExpr = outer

	next := p.Peek()
	if next.Kind() != RightBracket {
		p.report("']' expected", next.Loc())
		if expr != nil {
			loc.End = expr.Loc().End
		}
	} else {
		loc.End = p.Consume().Loc().End
	}

	return &BracketedExpression{expr, loc}
}
