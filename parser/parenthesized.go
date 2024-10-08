package parser

// An expression grouped between parentheses
type ParenthesizedExpression struct {
	Expr Expression
	loc  Loc
}

func (p ParenthesizedExpression) Loc() Loc {
	return p.loc
}

func (p ParenthesizedExpression) Type() ExpressionType {
	if p.Expr == nil {
		return Primitive{NIL}
	}
	if param, ok := p.Expr.(Param); ok {
		return Type{Object{map[string]ExpressionType{
			param.Identifier.Text(): param.Complement.Type(),
		}}}
	}
	return p.Expr.Type()
}

func (p ParenthesizedExpression) Unwrap() Expression {
	if expr, ok := p.Expr.(ParenthesizedExpression); ok {
		return expr.Unwrap()
	}
	return p.Expr
}

func (p *Parser) parseParenthesizedExpression() ParenthesizedExpression {
	loc := p.Consume().Loc() // LPAREN
	p.DiscardLineBreaks()
	next := p.Peek()
	if next.Kind() == RightParenthesis {
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
	if next.Kind() != RightParenthesis {
		p.report("')' expected", next.Loc())
	}
	loc.End = p.Consume().Loc().End
	return ParenthesizedExpression{expr, loc}
}

// unwrap parenthesized expressions
func Unwrap(expr Expression) Expression {
	if paren, ok := expr.(ParenthesizedExpression); ok {
		return paren.Unwrap()
	}
	return expr
}
