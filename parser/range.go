package parser

type RangeExpression struct {
	Left     Expression
	Right    Expression
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

func (r RangeExpression) Type() ExpressionType {
	var typing ExpressionType
	if r.Left != nil {
		typing = r.Left.Type()
	} else if r.Right != nil {
		typing = r.Right.Type()
	}
	return Range{typing}
}

func (p *Parser) parseRange() Expression {
	token := p.Peek()

	var left Expression
	if token.Kind() != InclusiveRange && token.Kind() != ExclusiveRange {
		left = BinaryExpression{}.Parse(p)
	}

	token = p.Peek()
	if token.Kind() != InclusiveRange && token.Kind() != ExclusiveRange {
		return left
	}
	operator := p.Consume()

	right := BinaryExpression{}.Parse(p)

	return &RangeExpression{left, right, operator}
}

func validateRangeExpression(p *Parser, r RangeExpression) {
	if r.Operator.Kind() == InclusiveRange && r.Right == nil {
		p.report(
			"Expected right operand with inclusive range operator '..='",
			r.Operator.Loc(),
		)
	}

	if r.Left != nil && r.Right != nil && !Match(r.Left.Type(), r.Right.Type()) {
		p.report("Left and right types don't match", r.Loc())
	}
}
