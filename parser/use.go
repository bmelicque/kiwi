package parser

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
	// TODO:
}

func (p *Parser) parseUseDirective() *UseDirective {
	start := p.Consume().Loc().Start // "use"
	star := false
	if p.Peek().Kind() == Mul {
		p.Consume()
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
