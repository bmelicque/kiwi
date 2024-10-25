package parser

// (param ParamType, otherParam OtherParamType)
type Params struct {
	Params []Param
	loc    Loc
}

func (p *Params) getChildren() []Node {
	children := make([]Node, len(p.Params))
	for i := range p.Params {
		children[i] = &p.Params[i]
	}
	return children
}

func (p Params) Loc() Loc { return p.loc }

// FIXME: object
func (p Params) Type() ExpressionType {
	types := make([]ExpressionType, len(p.Params))
	for i, element := range p.Params {
		types[i] = element.Type()
	}
	return Tuple{types}
}

func (params *Params) typeCheck(p *Parser) {
	for i := range params.Params {
		params.Params[i].Complement.typeCheck(p)
	}
}

// An identifier followed by a type expression
type Param struct {
	Identifier *Identifier
	Complement Expression // Type for params, value for arguments
	HasColon   bool
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
		return Primitive{UNKNOWN}
	}
	typing, ok := p.Complement.Type().(Type)
	if !ok {
		return Primitive{UNKNOWN}
	}
	return typing.Value
}

// TODO: rename this (taggedExpression?)
func (p *Parser) parseParam() Expression {
	expr := p.parseRange()
	identifier, ok := expr.(*Identifier)
	if !ok {
		return expr
	}
	colon := false
	if p.Peek().Kind() == Colon {
		p.Consume()
		colon = true
	}
	outer := p.allowEmptyExpr
	if !colon {
		p.allowEmptyExpr = true
	}
	expr = p.parseRange()
	p.allowEmptyExpr = outer
	if expr == nil {
		return identifier
	}
	return &Param{
		Identifier: identifier,
		Complement: expr,
		HasColon:   colon,
	}
}

func (param *Param) typeCheck(p *Parser) {
	param.Complement.typeCheck(p)
	isType := param.Complement.Type().Kind() == TYPE
	if param.HasColon && isType {
		p.report("Value expected", param.Complement.Loc())
	}
	if !param.HasColon && !isType {
		p.report("Type expected", param.Complement.Loc())
	}
}
