package parser

type RangeExpression struct {
	Left     Expression
	Right    Expression
	Operator Token
}

func (r *RangeExpression) getChildren() []Node {
	children := make([]Node, 0, 2)
	if r.Left != nil {
		children = append(children, r.Left)
	}
	if r.Right != nil {
		children = append(children, r.Right)
	}
	return children
}

func (r *RangeExpression) Loc() Loc {
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

func (r *RangeExpression) Type() ExpressionType {
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
		left = p.parseBinaryExpression()
	}

	token = p.Peek()
	if token.Kind() != InclusiveRange && token.Kind() != ExclusiveRange {
		return left
	}
	operator := p.Consume()

	var right Expression
	if operator.Kind() == ExclusiveRange {
		outer := p.allowEmptyExpr
		p.allowEmptyExpr = true
		right = p.parseBinaryExpression()
		p.allowEmptyExpr = outer
	} else {
		right = p.parseBinaryExpression()
	}
	return &RangeExpression{left, right, operator}
}

func (r *RangeExpression) typeCheck(p *Parser) {
	r.Left.typeCheck(p)
	r.Right.typeCheck(p)
	if r.Left != nil && r.Right != nil && !Match(r.Left.Type(), r.Right.Type()) {
		p.error(r, MismatchedTypes, r.Left.Type(), r.Right.Type())
	}
}
