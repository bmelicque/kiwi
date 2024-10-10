package parser

type TupleExpression struct {
	Elements []Expression
	typing   ExpressionType
}

func (t *TupleExpression) typeCheck(p *Parser) {
	// TODO:
}

func (t *TupleExpression) Loc() Loc {
	return Loc{
		Start: t.Elements[0].Loc().Start,
		End:   t.Elements[len(t.Elements)-1].Loc().End,
	}
}
func (t *TupleExpression) Type() ExpressionType { return t.typing }

// Wrap the expression in a tuple if not one
func makeTuple(expr Expression) *TupleExpression {
	tuple, ok := expr.(*TupleExpression)
	if ok {
		return tuple
	}
	return &TupleExpression{
		Elements: []Expression{expr},
		typing:   expr.Type(),
	}
}

func (p *Parser) parseTupleExpression() Expression {
	var elements []Expression
	outer := p.allowEmptyExpr
	p.allowEmptyExpr = true
	for p.Peek().Kind() != EOF {
		el := p.parseSumType()
		if el == nil {
			break
		}
		elements = append(elements, el)

		if p.Peek().Kind() != Comma {
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
	return &TupleExpression{elements, nil}
}
