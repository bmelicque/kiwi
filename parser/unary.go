package parser

type UnaryExpression struct {
	Operator Token
	Operand  Expression
}

func (u UnaryExpression) Loc() Loc {
	loc := u.Operator.Loc()
	if u.Operand != nil {
		loc.End = u.Operand.Loc().End
	}
	return loc
}

func (u UnaryExpression) Type() ExpressionType {
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
	Bracketed BracketedExpression
	Expr      Expression // Cannot be nil
}

func (l ListTypeExpression) Loc() Loc {
	loc := l.Bracketed.Loc()
	if l.Expr != nil {
		loc.End = l.Expr.Loc().End
	}
	return loc
}

func (l ListTypeExpression) Type() ExpressionType {
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
		if expr.Type().Kind() != TYPE {
			p.report("Type expected with question mark operator", expr.Loc())
		}
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
	// TODO: if brackets holds something, try to parse it as TypeParams and add them to scope?
	// TODO: if next is '(', try function?
	if p.Peek().Kind() == LeftParenthesis {

	}
	expr := parseInnerUnary(p)
	// FIXME: also FunctionTypeExpression
	function, ok := expr.(*FunctionExpression)
	if ok && function.TypeParams == nil {
		function.TypeParams = p.getValidatedTypeParams(*brackets)
		// TODO: validate function (here?)
		return function
	}
	list := ListTypeExpression{brackets, expr}
	validateListExpressionType(p, list)
	return &list
}

func validateListExpressionType(p *Parser, expr ListTypeExpression) {
	if expr.Bracketed.Expr != nil {
		p.report("No expression expected for list type", expr.Bracketed.Loc())
	}
	if expr.Expr != nil && expr.Expr.Type().Kind() != TYPE {
		p.report("Type expected", expr.Loc())
	}
}
