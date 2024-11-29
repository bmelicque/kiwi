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
	el := getIteratedElementType(expr.Right.Type())
	if el == (Unknown{}) {
		p.report("Invalid type, list or slice expected", expr.Right.Loc())
		return
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
func typeCheckForExpression(p *Parser, expr Expression) {
	if expr == nil {
		return
	}
	if _, ok := expr.Type().(Boolean); !ok {
		p.report("Boolean expected in loop condition", expr.Loc())
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
		p.report("Expression expected", operator.Loc())
	}
	right := p.parseExpression()
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
	case *Identifier:
	case *TupleExpression:
		expr.Left = getValidatedForInTuple(p, left)
	default:
		p.report("Invalid pattern", expr.Left.Loc())
		expr.Left = nil
	}
}

func getValidatedForInTuple(p *Parser, tuple *TupleExpression) *TupleExpression {
	if len(tuple.Elements) > 2 {
		p.report("Only 2 elements expected", tuple.Loc())
	}

	index, ok := tuple.Elements[0].(*Identifier)
	if !ok {
		p.report("Identifier expected", tuple.Elements[0].Loc())
	}
	value, ok := tuple.Elements[1].(*Identifier)
	if !ok {
		p.report("Identifier expected", tuple.Elements[1].Loc())
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
			p.report("No value expected", b.Value.Loc())
		}
		if t != (Nil{}) && !t.Extends(b.Value.Type()) {
			p.report("Type doesn't match the type inferred from first break", b.Value.Loc())
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
	}
	return Unknown{}
}
