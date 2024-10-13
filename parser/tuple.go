package parser

type TupleExpression struct {
	Elements []Expression
	typing   ExpressionType
}

func (t *TupleExpression) typeCheck(p *Parser) {
	for i := range t.Elements {
		t.Elements[i].typeCheck(p)
	}
	if len(t.Elements) == 0 {
		t.typing = Primitive{NIL}
		return
	}
	if len(t.Elements) == 1 {
		t.typing = t.Elements[0].Type()
		return
	}
	types := make([]ExpressionType, len(t.Elements))
	for i := range t.Elements {
		types[i] = t.Elements[i].Type()
	}
	t.typing = Tuple{types}
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
	if expr == nil {
		return &TupleExpression{nil, Primitive{NIL}}
	}
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
