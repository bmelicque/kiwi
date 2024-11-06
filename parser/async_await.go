package parser

type AsyncExpression struct {
	Keyword Token
	Call    *CallExpression
}

func (a *AsyncExpression) Loc() Loc {
	loc := a.Keyword.Loc()
	if a.Call != nil {
		loc.End = a.Call.Loc().End
	}
	return loc
}
func (a *AsyncExpression) getChildren() []Node {
	return a.Call.getChildren()
}
func (a *AsyncExpression) typeCheck(p *Parser) {
	if a.Call == nil {
		return
	}
	a.Call.typeCheck(p)
	if a.Call.Callee == nil {
		return
	}
	f, ok := a.Call.Callee.Type().(Function)
	if !ok {
		p.report("Function expected", a.Call.Loc())
		return
	}
	if !f.Async {
		p.report("'async' keyword has no effect in this expression", a.Loc())
	}
}
func (a *AsyncExpression) Type() ExpressionType {
	if a.Call == nil {
		return Unknown{}
	}
	return makePromise(a.Call.Type())
}

// Parse an async expression. Expects the next token to be 'async'.
func (p *Parser) parseAsyncExpression() *AsyncExpression {
	keyword := p.Consume() // AsyncKeyword
	expression := p.parseRange()
	call, ok := expression.(*CallExpression)
	if expression != nil && !ok {
		p.report("Call expression expected", expression.Loc())
	}
	return &AsyncExpression{keyword, call}
}

type AwaitExpression struct {
	Keyword Token
	Expr    Expression
}

func (a *AwaitExpression) Loc() Loc {
	loc := a.Keyword.Loc()
	if a.Expr != nil {
		loc.End = a.Expr.Loc().End
	}
	return loc
}
func (a *AwaitExpression) getChildren() []Node {
	return a.Expr.getChildren()
}
func (a *AwaitExpression) typeCheck(p *Parser) {
	if a.Expr == nil {
		return
	}
	a.Expr.typeCheck(p)
	alias, ok := a.Expr.Type().(TypeAlias)
	if !ok || alias.Name != "..." {
		p.report("Promise expected", a.Expr.Loc())
	}
}
func (a *AwaitExpression) Type() ExpressionType {
	if a.Expr == nil {
		return Unknown{}
	}
	alias, ok := a.Expr.Type().(TypeAlias)
	if !ok || alias.Name != "..." {
		return Unknown{}
	}
	t, _ := alias.Params[0].Value.build(nil, nil)
	return t
}

// Parse an AwaitExpression. Expects the next token to be 'await'.
func (p *Parser) parseAwaitExpression() *AwaitExpression {
	keyword := p.Consume() // AwaitKeyword
	expression := p.parseRange()
	return &AwaitExpression{keyword, expression}
}
