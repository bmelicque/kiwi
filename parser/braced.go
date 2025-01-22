package parser

// Expression between braces, such as `{value}`
type BracedExpression struct {
	Expr Expression
	loc  Loc
}

func (b *BracedExpression) getChildren() []Node {
	if b.Expr == nil {
		return []Node{}
	}
	return []Node{b.Expr}
}

func (b *BracedExpression) typeCheck(p *Parser) {
	if b.Expr == nil {
		return
	}
	b.Expr.typeCheck(p)
	b.Expr = makeTuple(b.Expr)
	for i, element := range b.Expr.(*TupleExpression).Elements {
		switch element := element.(type) {
		case *Identifier, *PropertyAccessExpression:
			if _, ok := element.Type().(Type); !ok {
				p.error(element, TypeExpected)
			}
		case *Param:
			if element.Identifier.IsPrivate() {
				p.error(element, MissingDefault)
			}
		case *Entry:
		default:
			p.error(element, InvalidPattern)
			b.Expr.(*TupleExpression).Elements[i] = nil
		}
	}
}

func (b *BracedExpression) Loc() Loc { return b.loc }
func (b *BracedExpression) Type() ExpressionType {
	o := newObject()
	for _, element := range b.Expr.(*TupleExpression).Elements {
		switch element := element.(type) {
		case *Identifier:
			if _, ok := element.Type().(Type); ok {
				o.addEmbedded(element.Text(), element.typing.(Type).Value)
			}
		case *PropertyAccessExpression:
			if _, ok := element.Type().(Type); ok {
				o.addEmbedded(element.Property.(*Identifier).Text(), element.typing.(Type).Value)
			}
		case *Param:
			var memberType ExpressionType
			if t, ok := element.Complement.Type().(Type); ok {
				memberType = t.Value
			} else {
				memberType = Invalid{}
			}
			if element.Identifier.IsPrivate() {
				// This case is reported as an error in the type-check phase.
				// However, to ensure non malfunction during import, this next line is necessary.
				o.addDefault(element.Identifier.Text(), memberType)
			} else {
				o.addMember(element.Identifier.Text(), memberType)
			}
		case *Entry:
			o.addDefault(element.Key.(*Identifier).Text(), element.Value.Type())
		}
	}
	return Type{o}
}
func (p *Parser) parseBracedExpression() *BracedExpression {
	if p.Peek().Kind() != LeftBrace {
		panic("'{' expected!")
	}
	loc := p.Consume().Loc()
	p.DiscardLineBreaks()

	outerMultiline := p.multiline
	outerEmpty := p.allowEmptyExpr
	p.multiline = true
	p.allowEmptyExpr = true
	expr := p.parseTupleExpression()
	p.multiline = outerMultiline
	p.allowEmptyExpr = outerEmpty

	p.DiscardLineBreaks()
	next := p.Peek()
	if next.Kind() != RightBrace {
		p.error(&Literal{next}, RightBraceExpected)
		if expr != nil {
			loc.End = expr.Loc().End
		}
	} else {
		loc.End = p.Consume().Loc().End
	}

	return &BracedExpression{expr, loc}
}
