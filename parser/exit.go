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
		p.error(value, UnexpectedExpression)
	}
	if operator == ThrowKeyword && value == nil {
		p.error(&Literal{p.Peek()}, ExpressionExpected)
	}

	statement := &Exit{keyword, value}
	inLoop := p.scope.in(LoopScope)
	inFunction := p.scope.in(FunctionScope)
	if operator == BreakKeyword && !inLoop {
		p.error(statement, IllegalBreak)
	}
	if operator == ContinueKeyword && !inLoop {
		p.error(statement, IllegalContinue)
	}
	if operator == ReturnKeyword && !inFunction {
		p.error(statement, IllegalReturn)
	}
	if operator == ThrowKeyword && !inFunction {
		p.error(statement, IllegalThrow)
	}
	return statement
}
