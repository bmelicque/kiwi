package parser

type ParenthesizedExpression struct {
	Expr Node
	loc  Loc
}

func (p ParenthesizedExpression) Loc() Loc {
	return p.loc
}
func (p ParenthesizedExpression) Unwrap() Node {
	if expr, ok := p.Expr.(ParenthesizedExpression); ok {
		return expr.Unwrap()
	}
	return p.Expr
}

func (p *Parser) parseParenthesizedExpression() ParenthesizedExpression {
	loc := p.Consume().Loc() // LPAREN
	p.DiscardLineBreaks()
	next := p.Peek()
	if next.Kind() == RPAREN {
		loc.End = p.Consume().Loc().End
		return ParenthesizedExpression{nil, loc}
	}

	outerBrace := p.allowBraceParsing
	outerMultiline := p.multiline
	p.allowBraceParsing = true
	p.multiline = true
	expr := p.parseTupleExpression()
	p.allowBraceParsing = outerBrace
	p.multiline = outerMultiline

	p.DiscardLineBreaks()
	next = p.Peek()
	if next.Kind() != RPAREN {
		p.report("')' expected", next.Loc())
	}
	loc.End = p.Consume().Loc().End
	return ParenthesizedExpression{expr, loc}
}

// unwrap parenthesized expressions
func Unwrap(node Node) Node {
	if paren, ok := node.(ParenthesizedExpression); ok {
		return paren.Unwrap()
	}
	return node
}
