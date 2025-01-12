package parser

import (
	"path/filepath"
)

type UseDirective struct {
	Names  Expression // *Identifier | *TupleExpression{[]*Identifier}
	Star   bool       // is 'use * as XXX from YYY'
	Source *Literal
	start  Position
}

func (u *UseDirective) Loc() Loc {
	return Loc{
		Start: u.start,
		End:   u.Source.Loc().End, // FIXME: can be nil
	}
}
func (u *UseDirective) getChildren() []Node { return []Node{} }
func (u *UseDirective) typeCheck(p *Parser) {
	module := typeCheckModule(p, u.Source)
	if u.Star {
		id, ok := u.Names.(*Identifier)
		if ok {
			addVariableToScope(p, id, module)
		}
	} else {
		declareUseNames(p, module, u.Names)
	}
}
func typeCheckModule(p *Parser, source Expression) ExpressionType {
	l, ok := source.(*Literal)
	if !ok {
		return Unknown{}
	}
	path := l.Text()
	path = path[1 : len(path)-1]
	var module Module
	if IsLocalPath(path) {
		path = filepath.Join(filepath.Dir(p.filePath), path)
		module, ok = filesExports[path]
	} else {
		module, ok = getLib(path)
	}
	if !ok {
		p.error(l, CannotResolvePath)
		return Unknown{}
	}
	return module
}

// 'names' should be either *Identifier or *TupleExpression{*Identifier}.
func declareUseNames(p *Parser, module ExpressionType, names Expression) {
	tuple := makeTuple(names)
	for _, el := range tuple.Elements {
		id := el.(*Identifier)
		switch module := module.(type) {
		case Unknown:
			addVariableToScope(p, id, module)
		case Module:
			t, ok := module.GetOwned(id.Text())
			if !ok {
				p.error(id, NotInModule, id.Text())
				t = Unknown{}
			}
			addVariableToScope(p, id, t)
		}
	}
}

func (p *Parser) parseUseDirective() *UseDirective {
	start := p.Consume().Loc().Start // "use"
	star := false
	if p.Peek().Kind() == Mul {
		p.Consume()
		star = true
		if p.Peek().Kind() != AsKeyword {
			recover(p, AsKeyword)
		} else {
			p.Consume()
		}
	}
	names := p.parseExpression()
	if p.Peek().Kind() != FromKeyword {
		recover(p, FromKeyword)
	} else {
		p.Consume()
	}
	expr := p.parseExpression()
	source, ok := expr.(*Literal)
	if !ok {
		p.error(expr, StringLiteralExpected)
	}
	u := &UseDirective{
		Names:  names,
		Star:   star,
		Source: source,
		start:  start,
	}
	validateUseDirective(p, u)
	return u
}

func validateUseDirective(p *Parser, u *UseDirective) {
	validateUseDirectiveNames(p, u)
	if _, ok := u.Names.(*Identifier); !ok && u.Star {
		p.error(u.Names, IdentifierExpected)
	}

	// TODO: resolve path
}

func validateUseDirectiveNames(p *Parser, u *UseDirective) {
	switch names := u.Names.(type) {
	case *Identifier:
	case *TupleExpression:
		for i, el := range names.Elements {
			if _, ok := el.(*Identifier); !ok {
				p.error(el, IdentifierExpected)
				names.Elements[i] = nil
			}
		}
	default:
		p.error(names, IdentifierExpected)
		u.Names = nil
	}
}
