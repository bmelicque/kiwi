package parser

import "fmt"

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
		p.report("Invalid pattern, expected 'Type{identifier}'", pattern.Loc())
		return
	}
	callee, ok := instance.Typing.(*Identifier)
	if !ok || !callee.IsType() {
		p.report("Type identifier expected", instance.Typing.Loc())
		return
	}
	v, ok := p.scope.Find(callee.Text())
	if !ok {
		p.report(
			fmt.Sprintf("Cannot find type '%v'", callee.Text()),
			callee.Loc(),
		)
		return
	}
	alias, ok := v.Typing.(TypeAlias)
	if !ok || alias.implements(trait) {
		msg := fmt.Sprintf("Type '%v' doesn't implement this trait", callee.Text())
		p.report(msg, callee.Loc())
		return
	}

	elements := instance.Args.Expr.(*TupleExpression).Elements
	if len(elements) != 1 {
		p.report("Only 1 argument expected", instance.Args.loc)
		return
	}
	identifier, ok := elements[0].(*Identifier)
	if !ok {
		p.report("Invalid pattern, expected 'Type(identifier)'", elements[0].Loc())
		return
	}
	p.scope.Add(identifier.Text(), identifier.Loc(), alias)
}
