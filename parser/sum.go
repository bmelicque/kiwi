package parser

type SumType struct {
	Members []Node
	start   Position
}

func (s SumType) Loc() Loc {
	return Loc{
		Start: s.start,
		End:   s.Members[len(s.Members)-1].Loc().End,
	}
}

func (p *Parser) parseSumType() Node {
	if p.Peek().Kind() != BinaryOr {
		return p.parseTypedExpression()
	}

	start := p.Peek().Loc().Start
	members := []Node{}
	for p.Peek().Kind() == BinaryOr {
		p.Consume()
		members = append(members, p.parseTypedExpression())
		handleSumTypeBadTokens(p)
		p.DiscardLineBreaks()
	}
	return SumType{Members: members, start: start}
}

func handleSumTypeBadTokens(p *Parser) {
	err := false
	var start, end Position
	for p.Peek().Kind() != EOL && p.Peek().Kind() != EOF && p.Peek().Kind() != BinaryOr {
		token := p.Consume()
		if !err {
			err = true
			start = token.Loc().Start
		}
		end = token.Loc().End
	}
	if err {
		p.report("EOL or '|' expected", Loc{Start: start, End: end})
	}
}
