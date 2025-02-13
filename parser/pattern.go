package parser

func (p *Parser) typeCheckPattern(pattern Expression, matched ExpressionType) {
	switch matched := matched.(type) {
	case Sum:
		validateSumPattern(p, pattern, matched)
	case Trait:
		validateTraitPattern(p, pattern, matched)
	}
}

func validateSumPattern(p *Parser, pattern Expression, sum Sum) {
	param, ok := pattern.(*Param)
	if !ok {
		p.error(pattern, InvalidPattern)
		return
	}
	typing, ok := param.Complement.(*Identifier)
	if !ok || !typing.IsType() {
		p.error(param.Complement, TypeIdentifierExpected)
		return
	}
	i := sum.getMember(typing.Text())
	if i == (Invalid{}) {
		p.error(typing, TypeDoesNotImplement, i)
	}
	p.scope.Add(param.Identifier.Text(), param.Identifier.Loc(), i)
}

func validateTraitPattern(p *Parser, pattern Expression, trait Trait) {
	param, ok := pattern.(*Param)
	if !ok {
		p.error(pattern, InvalidPattern)
		return
	}
	typing, ok := param.Complement.(*Identifier)
	if !ok || !typing.IsType() {
		p.error(param.Complement, TypeIdentifierExpected)
		return
	}
	typing.typeCheck(p)
	v, ok := p.scope.Find(typing.Text())
	if !ok {
		p.error(typing, CannotFind, typing.Text())
		return
	}
	alias, ok := v.Typing.(TypeAlias)
	if !ok || alias.Implements(trait) {
		p.error(typing, TypeDoesNotImplement, typing.Type())
		return
	}

	p.scope.Add(param.Identifier.Text(), param.Identifier.Loc(), alias)
}
