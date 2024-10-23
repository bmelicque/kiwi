package parser

type UnaryExpression struct {
	Operator Token
	Operand  Expression
}

func (u *UnaryExpression) Walk(cb func(Node), skip func(Node) bool) {
	if skip(u) {
		return
	}
	cb(u)
	if u.Operand != nil {
		u.Operand.Walk(cb, skip)
	}
}

func (u *UnaryExpression) typeCheck(p *Parser) {
	u.Operand.typeCheck(p)
	switch u.Operator.Kind() {
	case QuestionMark:
		if u.Operand.Type().Kind() != TYPE {
			p.report("Type expected with question mark operator", u.Operand.Loc())
		}
	}
}

func (u *UnaryExpression) Loc() Loc {
	loc := u.Operator.Loc()
	if u.Operand != nil {
		loc.End = u.Operand.Loc().End
	}
	return loc
}

func (u *UnaryExpression) Type() ExpressionType {
	switch u.Operator.Kind() {
	case QuestionMark:
		t := u.Operand.Type()
		if ty, ok := t.(Type); ok {
			t = ty.Value
		}
		return Type{makeOptionType(t)}
	default:
		return Primitive{UNKNOWN}
	}
}

type ListTypeExpression struct {
	Bracketed *BracketedExpression
	Expr      Expression // Cannot be nil
}

func (l *ListTypeExpression) Walk(cb func(Node), skip func(Node) bool) {
	if skip(l) {
		return
	}
	cb(l)
	if l.Expr != nil {
		l.Expr.Walk(cb, skip)
	}
}

func (l *ListTypeExpression) typeCheck(p *Parser) {
	l.Expr.typeCheck(p)
	if l.Expr != nil && l.Expr.Type().Kind() != TYPE {
		p.report("Type expected", l.Loc())
	}
}

func (l *ListTypeExpression) Loc() Loc {
	loc := l.Bracketed.Loc()
	if l.Expr != nil {
		loc.End = l.Expr.Loc().End
	}
	return loc
}

func (l *ListTypeExpression) Type() ExpressionType {
	t, ok := l.Expr.Type().(Type)
	if !ok {
		return Type{List{Primitive{UNKNOWN}}}
	}
	return Type{List{t.Value}}
}

func (p *Parser) parseUnaryExpression() Expression {
	switch p.Peek().Kind() {
	case QuestionMark:
		token := p.Consume()
		expr := parseInnerUnary(p)
		return &UnaryExpression{token, expr}
	case LeftBracket:
		return parseListTypeExpression(p)
	default:
		return p.parseAccessExpression()
	}
}

func parseInnerUnary(p *Parser) Expression {
	memBrace := p.allowBraceParsing
	memCall := p.allowCallExpr
	p.allowBraceParsing = false
	p.allowCallExpr = false
	expr := p.parseUnaryExpression()
	p.allowBraceParsing = memBrace
	p.allowCallExpr = memCall
	return expr
}

func parseListTypeExpression(p *Parser) Expression {
	brackets := p.parseBracketedExpression()
	if p.Peek().Kind() == LeftParenthesis {
		return p.parseFunctionExpression(brackets)
	}
	expr := parseInnerUnary(p)
	if brackets != nil && brackets.Expr != nil {
		p.report("No expression expected for list type", brackets.Loc())
	}
	return &ListTypeExpression{brackets, expr}
}
