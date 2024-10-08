package parser

type TupleExpression struct {
	Elements []Node
}

func (t TupleExpression) Loc() Loc {
	return Loc{
		Start: t.Elements[0].Loc().Start,
		End:   t.Elements[len(t.Elements)-1].Loc().End,
	}
}

func (p *Parser) parseTupleExpression() Node {
	var elements []Node
	outer := p.allowEmptyExpr
	p.allowEmptyExpr = true
	for p.Peek().Kind() != EOF {
		el := p.parseSumType()
		if el == nil {
			break
		}
		elements = append(elements, el)

		if p.Peek().Kind() != COMMA {
			break
		}
		p.Consume()

		if p.multiline {
			p.DiscardLineBreaks()
		}
	}
	p.allowEmptyExpr = outer

	if len(elements) == 0 {
		return nil
	}
	if len(elements) == 1 {
		return elements[0]
	}
	return TupleExpression{elements}
}
