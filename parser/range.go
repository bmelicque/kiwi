package parser

type RangeExpression struct {
	Left     Node
	Right    Node
	Operator Token
}

func (r RangeExpression) Loc() Loc {
	var loc Loc
	if r.Left != nil {
		loc.Start = r.Left.Loc().Start
	} else {
		loc.Start = r.Operator.Loc().Start
	}
	if r.Right != nil {
		loc.End = r.Right.Loc().End
	} else {
		loc.End = r.Operator.Loc().End
	}
	return loc
}

func ParseRange(p *Parser) Node {
	token := p.Peek()

	var left Node
	if token.Kind() != RANGE_INCLUSIVE && token.Kind() != RANGE_EXCLUSIVE {
		left = BinaryExpression{}.Parse(p)
	}

	token = p.Peek()
	if token.Kind() != RANGE_INCLUSIVE && token.Kind() != RANGE_EXCLUSIVE {
		return left
	}
	operator := p.Consume()

	right := BinaryExpression{}.Parse(p)

	return RangeExpression{left, right, operator}
}
