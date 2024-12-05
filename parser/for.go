package parser

type ForExpression struct {
	Keyword Token
	Expr    Expression // Expression (boolean)
	Body    *Block
	typing  ExpressionType
}

func (f *ForExpression) getChildren() []Node {
	children := []Node{}
	if f.Expr != nil {
		children = append(children, f.Expr)
	}
	if f.Body != nil {
		children = append(children, f.Body)
	}
	return children
}

func (f *ForExpression) typeCheck(p *Parser) {
	p.pushScope(NewScope(LoopScope))
	defer p.dropScope()

	binary, ok := f.Expr.(*BinaryExpression)
	if ok && binary.Operator.Kind() == InKeyword {
		typeCheckForInExpression(p, binary)
	} else {
		typeCheckForExpression(p, f.Expr)
	}
	f.Body.typeCheck(p)
	f.typing = getLoopType(p, f.Body)
}
func typeCheckForInExpression(p *Parser, expr *BinaryExpression) {
	var el ExpressionType = Unknown{}
	if expr.Right != nil {
		expr.Right.typeCheck(p)
		checkExplicitRange(p, expr.Right)
		el = getIteratedElementType(expr.Right.Type())
		if el == (Unknown{}) {
			p.error(expr.Right, IterableExpected, expr.Right.Type())
		}
	}
	switch pattern := expr.Left.(type) {
	case *Identifier:
		p.scope.Add(pattern.Text(), pattern.Loc(), el)
	case *TupleExpression:
		element := pattern.Elements[0].(*Identifier)
		if element != nil {
			p.scope.Add(element.Text(), element.Loc(), el)
		}
		index := pattern.Elements[1].(*Identifier)
		if index != nil {
			p.scope.Add(index.Text(), index.Loc(), Number{})
		}
	}
}
func checkExplicitRange(p *Parser, expr Expression) {
	_, ok := expr.Type().(Range)
	if !ok {
		return
	}
	_, ok = Unwrap(expr).(*RangeExpression)
	if !ok {
		p.error(expr, RangeExpected)
	}
}
func typeCheckForExpression(p *Parser, expr Expression) {
	if expr == nil {
		return
	}
	if _, ok := expr.Type().(Boolean); !ok {
		p.error(expr, BooleanExpected, expr.Type())
	}
}

func (f *ForExpression) Loc() Loc {
	loc := f.Keyword.Loc()
	if f.Body != nil {
		loc.End = f.Body.Loc().End
	} else if f.Expr != nil {
		loc.End = f.Expr.Loc().End
	}
	return loc
}
func (f *ForExpression) Type() ExpressionType { return f.typing }

func (p *Parser) parseForExpression() *ForExpression {
	p.pushScope(NewScope(LoopScope))
	defer p.dropScope()

	keyword := p.Consume()
	expr := parseInExpression(p)
	body := p.parseBlock()

	return &ForExpression{keyword, expr, body, nil}
}

func parseInExpression(p *Parser) Expression {
	brace := p.allowBraceParsing
	empty := p.allowEmptyExpr
	p.allowBraceParsing = false
	p.allowEmptyExpr = true
	defer func() { p.allowBraceParsing = brace }()
	expr := p.parseExpression()
	p.allowEmptyExpr = empty

	if p.Peek().Kind() != InKeyword {
		return expr
	}
	operator := p.Consume()
	if expr == nil {
		p.error(&Literal{operator}, ExpressionExpected)
	}
	right := p.parseRange()
	in := &BinaryExpression{
		Left:     expr,
		Right:    right,
		Operator: operator,
	}
	validateInExpression(p, in)
	return in
}

func validateInExpression(p *Parser, expr *BinaryExpression) {
	switch left := expr.Left.(type) {
	case nil, *Identifier:
	case *TupleExpression:
		expr.Left = getValidatedForInTuple(p, left)
	default:
		p.error(expr.Left, InvalidPattern)
		expr.Left = nil
	}
}

func getValidatedForInTuple(p *Parser, tuple *TupleExpression) *TupleExpression {
	if len(tuple.Elements) > 2 {
		p.error(tuple, TooManyElements, "at most 2", len(tuple.Elements))
	}

	index, ok := tuple.Elements[0].(*Identifier)
	if !ok {
		p.error(tuple.Elements[0], IdentifierExpected)
	}
	value, ok := tuple.Elements[1].(*Identifier)
	if !ok {
		p.error(tuple.Elements[1], IdentifierExpected)
	}
	return &TupleExpression{Elements: []Expression{index, value}}
}

func getLoopType(p *Parser, body *Block) ExpressionType {
	breaks := findBreakStatements(body)
	if len(breaks) == 0 {
		return Nil{}
	}
	var t ExpressionType
	if breaks[0].Value != nil {
		t = breaks[0].Value.Type()
	} else {
		t = Nil{}
	}
	for _, b := range breaks[1:] {
		if t == (Nil{}) && b.Value != nil {
			p.error(b.Value, CannotAssignType, t, b.Value.Type())
		}
		if t != (Nil{}) && !t.Extends(b.Value.Type()) {
			p.error(b.Value, CannotAssignType, t, b.Value.Type())
		}
	}
	return t
}

func findBreakStatements(body *Block) []*Exit {
	results := []*Exit{}
	Walk(body, func(n Node, skip func()) {
		if isFunctionExpression(n) || isForExpression(n) {
			skip()
		}
		if n, ok := n.(*Exit); ok && n.Operator.Kind() == BreakKeyword {
			results = append(results, n)
		}
	})
	return results
}

func isForExpression(n Node) bool {
	_, ok := n.(*ForExpression)
	return ok
}

// t might be a List or a Ref to a List.
// Return the type iterated on in a loop
func getIteratedElementType(t ExpressionType) ExpressionType {
	switch t := t.(type) {
	case List:
		return t.Element
	case Ref:
		if l, ok := t.To.(List); ok {
			return Ref{l.Element}
		}
	case Range:
		return t.operands
	}
	return Unknown{}
}
