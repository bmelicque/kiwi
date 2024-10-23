package parser

type Exit struct {
	Operator Token
	Value    Expression
}

func (e *Exit) Walk(cb func(Node), skip func(Node) bool) {
	if skip(e) {
		return
	}
	cb(e)
	if e.Value != nil {
		e.Value.Walk(cb, skip)
	}
}

func (e *Exit) typeCheck(p *Parser) {
	if e.Value != nil {
		e.Value.typeCheck(p)
	}
}

func (e *Exit) Loc() Loc {
	loc := e.Operator.Loc()
	if e.Value != nil {
		loc.End = e.Value.Loc().End
	}
	return loc
}

func (p *Parser) parseExit() *Exit {
	keyword := p.Consume()

	if p.Peek().Kind() == EOL {
		return &Exit{keyword, nil}
	}

	value := p.parseExpression()

	operator := keyword.Kind()
	if keyword.Kind() == ContinueKeyword && value != nil {
		p.report("No value expected after 'continue'", value.Loc())
	}

	statement := &Exit{keyword, value}
	if keyword.Kind() == ReturnKeyword && !p.scope.in(FunctionScope) {
		p.report("Cannot return outside of a function", statement.Loc())
	}
	if operator == BreakKeyword && !p.scope.in(LoopScope) {
		p.report("Cannot break outside of a loop", statement.Loc())
	}
	if operator == ContinueKeyword && !p.scope.in(LoopScope) {
		p.report("Cannot continue outside of a loop", statement.Loc())
	}
	return statement
}
