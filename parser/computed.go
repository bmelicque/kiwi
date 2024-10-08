package parser

// Expr[Property]
type ComputedAccessExpression struct {
	Expr     Expression
	Property *BracketedExpression
	typing   ExpressionType
}

func (c ComputedAccessExpression) Loc() Loc {
	return Loc{
		Start: c.Expr.Loc().Start,
		End:   c.Property.loc.End,
	}
}
func (c ComputedAccessExpression) Type() ExpressionType {
	return c.typing
}

func parseComputedAccessExpression(p *Parser, expr Expression) *ComputedAccessExpression {
	prop := p.parseBracketedExpression()
	typing := getComputedAccessType(p, expr, *prop)
	return &ComputedAccessExpression{
		Expr:     expr,
		Property: prop,
		typing:   typing,
	}
}

func getComputedAccessType(p *Parser, left Expression, right BracketedExpression) ExpressionType {
	tuple, ok := right.Expr.(*TupleExpression)
	if !ok {
		tuple = &TupleExpression{
			Elements: []Expression{right.Expr},
			typing:   right.Expr.Type(),
		}
	}
	switch t := left.Type().(type) {
	case Type:
		// Generics
		alias, ok := t.Value.(TypeAlias)
		if !ok {
			p.report("No type arguments expected", right.Expr.Loc())
			return Primitive{UNKNOWN}
		}
		p.pushScope(NewScope(ProgramScope))
		defer p.dropScope()
		params := append(alias.Params[:0:0], alias.Params...)
		p.addTypeArgsToScope(tuple, params)
		ref, _ := alias.Ref.build(p.scope, nil)
		return Type{TypeAlias{
			Name:   alias.Name,
			Params: params,
			Ref:    ref,
		}}
	case Function:
		// Generic function
		p.pushScope(NewScope(ProgramScope))
		defer p.dropScope()
		typeParams := append(t.TypeParams[:0:0], t.TypeParams...)
		p.addTypeArgsToScope(tuple, typeParams)
		params := make([]ExpressionType, len(typeParams))
		for i, param := range typeParams {
			params[i], _ = param.build(p.scope, nil)
		}
		returned, _ := t.Returned.build(p.scope, nil)
		return Function{
			TypeParams: typeParams,
			Params:     Tuple{params},
			Returned:   returned,
		}
	case List:
		if tuple.Type().Kind() != NUMBER {
			p.report("Number expected", tuple.Loc())
		}
		return t.Element
	default:
		return nil
	}
}
