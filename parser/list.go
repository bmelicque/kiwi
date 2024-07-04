package parser

import (
	"github.com/bmelicque/test-parser/tokenizer"
)

type ListExpression struct {
	elements []Expression
	loc      tokenizer.Loc
}

func (l ListExpression) Type(ctx *Scope) ExpressionType {
	if len(l.elements) == 0 {
		return List{Primitive{UNKNOWN}}
	}
	t := l.elements[0].Type(ctx)
	if t.Kind() == TYPE {
		return t
	}
	return List{t}
}

func (l ListExpression) Check(c *Checker) {
	var t ExpressionType
	for _, el := range l.elements {
		if el == nil {
			continue
		}
		el.Check(c)
		if t == nil {
			t = el.Type(c.scope)
		} else if !t.Extends(el.Type(c.scope)) {
			c.report("Types don't match", el.Loc())
		}
	}

}
func (l ListExpression) Emit(e *Emitter) {
	e.Write("[")
	length := len(l.elements)
	for i, el := range l.elements {
		el.Emit(e)
		if i != length-1 {
			e.Write(", ")
		}
	}
	e.Write("]")
}

func (l ListExpression) Loc() tokenizer.Loc { return l.loc }
func (l ListExpression) Parse(p *Parser) Expression {
	lbracket := p.tokenizer.Consume()
	l.loc.Start = lbracket.Loc().Start

	ParseList(p, tokenizer.RBRACKET, func() {
		l.elements = append(l.elements, ParseExpression(p))
	})

	next := p.tokenizer.Peek()
	if next.Kind() != tokenizer.RBRACKET {
		p.report("']' expected", next.Loc())
	}
	l.loc.End = next.Loc().End

	return l
}
