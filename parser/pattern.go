package parser

func (p *Parser) typeCheckPattern(pattern Expression, matched ExpressionType) {
	switch matched := matched.(type) {
	case Sum:
	case Trait:
		validateTraitPattern(p, pattern, matched)
	}
}

func validateTraitPattern(p *Parser, pattern Expression, trait Trait) {
	instance, ok := pattern.(*InstanceExpression)
	if !ok {
		p.error(pattern, InvalidPattern)
		return
	}
	callee, ok := instance.Typing.(*Identifier)
	if !ok || !callee.IsType() {
		p.error(instance.Typing, TypeIdentifierExpected)
		return
	}
	v, ok := p.scope.Find(callee.Text())
	if !ok {
		p.error(callee, CannotFind, callee.Text())
		return
	}
	alias, ok := v.Typing.(TypeAlias)
	if !ok || alias.Implements(trait) {
		p.error(callee, TypeDoesNotImplement, callee.Type())
		return
	}

	elements := instance.Args.Expr.(*TupleExpression).Elements
	if len(elements) != 1 {
		p.error(instance.Args, TooManyElements, 1, len(elements))
		return
	}
	identifier, ok := elements[0].(*Identifier)
	if !ok {
		p.error(elements[0], InvalidPattern)
		return
	}
	p.scope.Add(identifier.Text(), identifier.Loc(), alias)
}
