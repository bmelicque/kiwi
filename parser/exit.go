package parser

type Exit struct {
	Operator Token
	Value    Node
}

func (r Exit) Loc() Loc {
	loc := r.Operator.Loc()
	if r.Value != nil {
		loc.End = r.Value.Loc().End
	}
	return loc
}

func (p *Parser) parseExit() Exit {
	keyword := p.Consume()

	if p.Peek().Kind() == EOL {
		return Exit{keyword, nil}
	}

	value := ParseExpression(p)
	return Exit{keyword, value}
}
