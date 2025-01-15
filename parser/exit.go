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
	var value Expression
	if p.Peek().Kind() != EOL {
		value = p.parseExpression()
	}
	statement := &Exit{keyword, value}
	checkKeywordValueConsistency(p, statement)
	reportIllegalExit(p, statement)
	return statement
}

// check if presence/absence of value matches keyword
func checkKeywordValueConsistency(p *Parser, e *Exit) {
	operator := e.Operator.Kind()
	if operator == ContinueKeyword && e.Value != nil {
		p.error(e.Value, UnexpectedExpression)
	}
	if operator == ThrowKeyword && e.Value == nil {
		p.error(&Literal{p.Peek()}, ExpressionExpected)
	}
}

// Exit statements may only appear in certain contexts
// (e.g. you can only return in a function)
func reportIllegalExit(p *Parser, e *Exit) {
	operator := e.Operator.Kind()
	inLoop := p.scope.in(LoopScope)
	inFunction := p.scope.in(FunctionScope)
	if operator == BreakKeyword && !inLoop {
		p.error(e, IllegalBreak)
	}
	if operator == ContinueKeyword && !inLoop {
		p.error(e, IllegalContinue)
	}
	if operator == ReturnKeyword && !inFunction {
		p.error(e, IllegalReturn)
	}
	if operator == ThrowKeyword && !inFunction {
		p.error(e, IllegalThrow)
	}
}
