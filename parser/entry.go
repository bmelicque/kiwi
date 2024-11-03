package parser

// An key followed by a colon and a type expression.
// The key can be either an identifier, a [bracketed expression] or a literal.
type Entry struct {
	Key   Expression // *BracketedExpression | *Identifier | *Literal
	Value Expression
}

func (e *Entry) getChildren() []Node {
	children := make([]Node, 0, 2)
	if e.Key != nil {
		children = append(children, e.Key)
	}
	if e.Value != nil {
		children = append(children, e.Value)
	}
	return children
}
func (e *Entry) Loc() Loc {
	var start, end Position
	if e.Key != nil {
		start = e.Key.Loc().Start
	} else {
		start = e.Value.Loc().Start
	}
	if e.Value != nil {
		end = e.Value.Loc().End
	} else {
		end = e.Key.Loc().End
	}
	return Loc{start, end}
}
func (e *Entry) Type() ExpressionType {
	if e.Value == nil {
		return Unknown{}
	}
	return e.Value.Type()
}

// An identifier followed by a type expression
type Param struct {
	Identifier *Identifier
	Complement Expression
}

func (p *Param) getChildren() []Node {
	children := make([]Node, 0, 2)
	if p.Identifier != nil {
		children = append(children, p.Identifier)
	}
	if p.Complement != nil {
		children = append(children, p.Complement)
	}
	return children
}

func (p *Param) Loc() Loc {
	var start, end Position
	if p.Identifier != nil {
		start = p.Identifier.Loc().Start
	} else {
		start = p.Complement.Loc().Start
	}
	if p.Complement != nil {
		end = p.Complement.Loc().End
	} else {
		end = p.Identifier.Loc().End
	}
	return Loc{start, end}
}
func (p *Param) Type() ExpressionType {
	if p.Complement == nil {
		return Unknown{}
	}
	typing, ok := p.Complement.Type().(Type)
	if !ok {
		return Unknown{}
	}
	return typing.Value
}

func (p *Parser) parseTaggedExpression() Expression {
	expr := p.parseRange()
	if p.Peek().Kind() == Colon {
		return parseEntry(p, expr)
	}
	if identifier, ok := expr.(*Identifier); ok {
		return parseParam(p, identifier)
	}
	return expr
}

func parseEntry(p *Parser, expr Expression) Expression {
	if p.preventColon {
		return expr
	}
	p.Consume()
	switch expr.(type) {
	case *BracketedExpression, *Identifier, *Literal:
	default:
		p.report(
			"Invalid expression before ':': identifier, literal or brackets expected",
			expr.Loc(),
		)
		expr = nil
	}
	complement := p.parseRange()
	return &Entry{
		Key:   expr,
		Value: complement,
	}
}

func parseParam(p *Parser, identifier *Identifier) Expression {
	outer := p.allowEmptyExpr
	p.allowEmptyExpr = true
	expr := p.parseRange()
	p.allowEmptyExpr = outer
	if expr == nil {
		return identifier
	}
	return &Param{
		Identifier: identifier,
		Complement: expr,
	}
}

func (param *Param) typeCheck(p *Parser) {
	param.Complement.typeCheck(p)
	if param.Complement.Type().Kind() != TYPE {
		p.report("Type expected", param.Complement.Loc())
	}
}

func (e *Entry) typeCheck(p *Parser) {
	e.Value.typeCheck(p)
	if b, ok := e.Key.(*BracketedExpression); ok {
		b.typeCheck(p)
	}
	if e.Value.Type().Kind() == TYPE {
		p.report("Value expected", e.Value.Loc())
	}
}
