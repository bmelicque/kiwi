package parser

type FunctionExpression struct {
	TypeParams *BracketedExpression
	Params     *ParenthesizedExpression
	Operator   Token // -> or =>
	Expr       Node  // return value for '->', return type for '=>'
	Body       *Block
}

func (f FunctionExpression) Loc() Loc {
	loc := Loc{Start: f.Params.Loc().Start, End: Position{}}
	if f.Body == nil {
		loc.End = f.Expr.Loc().End
	} else {
		loc.End = f.Body.Loc().End
	}
	return loc
}

func (p *Parser) parseFunctionExpression() Node {
	paren := p.parseParenthesizedExpression()

	next := p.Peek()
	if next.Kind() != SLIM_ARR && next.Kind() != FAT_ARR {
		return paren
	}
	operator := p.Consume()

	next = p.Peek()
	if next.Kind() == LBRACE {
		p.report("Expression expected", next.Loc())
	}

	var expr Node
	if operator.Kind() == FAT_ARR {
		old := p.allowBraceParsing
		p.allowBraceParsing = false
		expr = ParseRange(p)
		p.allowBraceParsing = old
	} else {
		expr = ParseRange(p)
	}
	res := FunctionExpression{nil, &paren, operator, expr, nil}
	if operator.Kind() == FAT_ARR {
		res.Body = p.parseBlock()
	}
	return res
}
