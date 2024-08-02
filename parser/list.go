package parser

import (
	"github.com/bmelicque/test-parser/tokenizer"
)

type ListExpression struct {
	Elements []Expression
	loc      tokenizer.Loc
}

func (l ListExpression) Type() ExpressionType {
	if len(l.Elements) == 0 {
		return List{Primitive{UNKNOWN}}
	}
	t := l.Elements[0].Type()
	if t.Kind() == TYPE {
		return t
	}
	return List{t}
}

func (l ListExpression) Check(c *Checker) {
	var t ExpressionType
	for _, el := range l.Elements {
		if el == nil {
			continue
		}
		el.Check(c)
		if t == nil {
			t = el.Type()
		} else if !t.Extends(el.Type()) {
			c.report("Types don't match", el.Loc())
		}
	}

}

func (l ListExpression) Loc() tokenizer.Loc { return l.loc }
func (l ListExpression) Parse(p *Parser) Expression {
	lbracket := p.tokenizer.Consume()
	l.loc.Start = lbracket.Loc().Start

	ParseList(p, tokenizer.RBRACKET, func() {
		l.Elements = append(l.Elements, ParseExpression(p))
	})

	next := p.tokenizer.Peek()
	if next.Kind() != tokenizer.RBRACKET {
		p.report("']' expected", next.Loc())
	}
	l.loc.End = next.Loc().End

	return l
}
