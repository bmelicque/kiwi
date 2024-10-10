package parser

import "fmt"

func (p *Parser) validatePattern(pattern Expression, matched ExpressionType) {
	switch matched := matched.(type) {
	case Sum:
	case Trait:
		validateTraitPattern(p, pattern, matched)
	}
}

func validateTraitPattern(p *Parser, pattern Expression, trait Trait) {
	call, ok := pattern.(*CallExpression)
	if !ok {
		p.report("Invalid pattern, expected 'Type(identifier)'", pattern.Loc())
		return
	}
	callee, ok := call.Callee.(*Identifier)
	if !ok || !callee.isType {
		p.report("Type identifier expected", call.Callee.Loc())
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
	alias, ok := v.typing.(TypeAlias)
	if !ok || alias.implements(trait) {
		msg := fmt.Sprintf("Type '%v' doesn't implement this trait", callee.Text())
		p.report(msg, callee.Loc())
		return
	}

	if len(call.Args.Params) != 1 {
		p.report("Only 1 argument expected", call.Args.loc)
		return
	}
	identifier := call.Args.Params[0].Identifier
	p.scope.Add(identifier.Text(), identifier.Loc(), alias)
}
