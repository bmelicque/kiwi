package parser

// Expr[Property]
type ComputedAccessExpression struct {
	Expr     Expression
	Property *BracketedExpression
	typing   ExpressionType
}

func (c *ComputedAccessExpression) getChildren() []Node {
	return []Node{c.Expr, c.Property}
}

func (c *ComputedAccessExpression) Loc() Loc {
	return Loc{
		Start: c.Expr.Loc().Start,
		End:   c.Property.loc.End,
	}
}
func (c *ComputedAccessExpression) Type() ExpressionType {
	return c.typing
}

func parseComputedAccessExpression(p *Parser, expr Expression) *ComputedAccessExpression {
	prop := p.parseBracketedExpression()
	return &ComputedAccessExpression{Expr: expr, Property: prop}
}

func (expr *ComputedAccessExpression) typeCheck(p *Parser) {
	expr.Expr.typeCheck(p)
	switch t := expr.Expr.Type().(type) {
	case Type:
		typeCheckGenericType(p, expr)
	case TypeAlias:
		typeCheckMapAccess(p, expr)
	case Function:
		typeCheckGenericFunction(p, expr)
	case List:
		if _, ok := expr.Property.Expr.Type().(Number); !ok {
			p.report("Number expected", expr.Property.loc)
		}
		expr.typing = t.Element
	}
}

func typeCheckMapAccess(p *Parser, expr *ComputedAccessExpression) {
	expr.Property.typeCheck(p)
	a := expr.Expr.Type().(TypeAlias)
	if a.Name != "Map" {
		expr.typing = Unknown{}
		return
	}
	b, _ := a.Ref.build(p.scope, nil)
	m := b.(Map)
	if expr.Property.Expr != nil && !m.Key.Extends(expr.Property.Expr.Type()) {
		p.report("Type doesn't match expected key type", expr.Property.loc)
	}
	expr.typing = makeOptionType(m.Value)
}

func typeCheckGenericType(p *Parser, expr *ComputedAccessExpression) {
	alias, ok := expr.Expr.Type().(Type).Value.(TypeAlias)
	if !ok {
		p.report("No type arguments expected for this type", expr.Property.loc)
		expr.typing = Unknown{}
		return
	}

	p.pushScope(NewScope(ProgramScope))
	defer p.dropScope()

	params := append(alias.Params[:0:0], alias.Params...)
	typeCheckTypeArgs(p, makeTuple(expr.Property.Expr), params)
	ref, _ := alias.Ref.build(p.scope, nil)
	expr.typing = Type{TypeAlias{
		Name:   alias.Name,
		Params: params,
		Ref:    ref,
	}}
}

func typeCheckGenericFunction(p *Parser, expr *ComputedAccessExpression) {
	p.pushScope(NewScope(ProgramScope))
	defer p.dropScope()

	t := expr.Expr.Type().(Function)
	typeParams := append(t.TypeParams[:0:0], t.TypeParams...)
	typeCheckTypeArgs(p, makeTuple(expr.Property.Expr), typeParams)
	params := make([]ExpressionType, len(typeParams))
	for i, param := range typeParams {
		params[i], _ = param.build(p.scope, nil)
	}
	returned, _ := t.Returned.build(p.scope, nil)
	expr.typing = Function{
		TypeParams: typeParams,
		Params:     &Tuple{params},
		Returned:   returned,
	}
}
