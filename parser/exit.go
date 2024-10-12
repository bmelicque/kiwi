package parser

type Exit struct {
	Operator Token
	Value    Expression
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

	value := ParseExpression(p)
	return &Exit{keyword, value}
}
