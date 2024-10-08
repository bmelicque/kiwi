package parser

type UnaryExpression struct {
	Operator Token
	Operand  Node
}

func (u UnaryExpression) Loc() Loc {
	loc := u.Operator.Loc()
	if u.Operand != nil {
		loc.End = u.Operand.Loc().End
	}
	return loc
}

type ListTypeExpression struct {
	Bracketed BracketedExpression
	Type      Node // Cannot be nil
}

func (l ListTypeExpression) Loc() Loc {
	end := l.Bracketed.Loc().End
	if l.Type != nil {
		end = l.Type.Loc().End
	}
	return Loc{Start: l.Bracketed.loc.Start, End: end}
}

func (p *Parser) parseUnaryExpression() Node {
	switch p.Peek().Kind() {
	case QuestionMark:
		token := p.Consume()
		expr := parseInnerUnary(p)
		return UnaryExpression{token, expr}
	case LeftBracket:
		brackets := p.parseBracketedExpression()
		expr := parseInnerUnary(p)
		if function, ok := expr.(FunctionExpression); ok {
			function.TypeParams = &brackets
			return function
		}
		return ListTypeExpression{brackets, expr}
	default:
		return p.parseAccessExpression()
	}
}

func parseInnerUnary(p *Parser) Node {
	memBrace := p.allowBraceParsing
	memCall := p.allowCallExpr
	p.allowBraceParsing = false
	p.allowCallExpr = false
	expr := p.parseUnaryExpression()
	p.allowBraceParsing = memBrace
	p.allowCallExpr = memCall
	return expr
}
