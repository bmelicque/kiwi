package parser

type Exit struct {
	Operator Token
	Value    Expression
}

func (e *Exit) getChildren() []Node {
	if e.Value == nil {
		return []Node{}
	}
	return []Node{e.Value}
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
	if operator == ContinueKeyword && value != nil {
		p.report("No value expected after 'continue'", value.Loc())
	}
	if operator == ThrowKeyword && value == nil {
		p.report("Value expected after 'throw'", keyword.Loc())
	}

	statement := &Exit{keyword, value}
	inLoop := p.scope.in(LoopScope)
	inFunction := p.scope.in(FunctionScope)
	if operator == BreakKeyword && !inLoop {
		p.report("Cannot break outside of a loop", statement.Loc())
	}
	if operator == ContinueKeyword && !inLoop {
		p.report("Cannot continue outside of a loop", statement.Loc())
	}
	if operator == ReturnKeyword && !inFunction {
		p.report("Cannot return outside of a function", statement.Loc())
	}
	if operator == ThrowKeyword && !inFunction {
		p.report("Cannot throw outside of a function", statement.Loc())
	}
	return statement
}
