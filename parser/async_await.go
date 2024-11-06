package parser

type AsyncExpression struct {
	Keyword Token
	Expr    Expression
}

func (a *AsyncExpression) Loc() Loc {
	loc := a.Keyword.Loc()
	if a.Expr != nil {
		loc.End = a.Expr.Loc().End
	}
	return loc
}
func (a *AsyncExpression) getChildren() []Node {
	return a.Expr.getChildren()
}
func (a *AsyncExpression) typeCheck(p *Parser) {
	a.Expr.typeCheck(p)
	call, ok := a.Expr.(*CallExpression)
	if !ok {
		p.report("Call expression expected", a.Expr.Loc())
		return
	}
	if call.Callee == nil {
		return
	}
	f, ok := call.Callee.Type().(Function)
	if !ok {
		p.report("Function expected", a.Expr.Loc())
		return
	}
	if !f.Async {
		p.report("'async' keyword has no effect in this expression", a.Loc())
	}
}
func (a *AsyncExpression) Type() ExpressionType {
	if a.Expr == nil {
		return Unknown{}
	}
	return makePromise(a.Expr.Type())
}

// Parse an async expression. Expects the next token to be 'async'.
func (p *Parser) parseAsyncExpression() *AsyncExpression {
	keyword := p.Consume() // AsyncKeyword
	expression := p.parseRange()
	return &AsyncExpression{keyword, expression}
}
