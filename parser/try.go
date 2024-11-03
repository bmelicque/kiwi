package parser

type TryExpression struct {
	Keyword Token
	Expr    Expression
}

func (t *TryExpression) getChildren() []Node {
	if t.Expr == nil {
		return []Node{}
	}
	return []Node{t.Expr}
}

func (t *TryExpression) Loc() Loc {
	loc := t.Keyword.Loc()
	if t.Expr != nil {
		loc.End = t.Expr.Loc().End
	}
	return loc
}

func (t *TryExpression) Type() ExpressionType {
	alias, ok := t.Expr.Type().(TypeAlias)
	if !ok || alias.Name != "Result" {
		return Unknown{}
	}
	return alias.Ref.(Sum).getMember("Ok")
}

func (t *TryExpression) typeCheck(p *Parser) {
	if t.Expr == nil {
		return
	}
	t.Expr.typeCheck(p)
	alias, ok := t.Expr.Type().(TypeAlias)
	if !ok || alias.Name != "Result" {
		p.report("Result type expected", t.Expr.Loc())
	}
}

func (p *Parser) parseTryExpression() *TryExpression {
	keyword := p.Consume() // try
	expr := p.parseExpression()
	return &TryExpression{
		Keyword: keyword,
		Expr:    expr,
	}
}
