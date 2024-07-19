package parser

import (
	"github.com/bmelicque/test-parser/tokenizer"
)

type TupleExpression struct {
	Elements []Expression
	loc      tokenizer.Loc
}

func (t TupleExpression) Type(ctx *Scope) ExpressionType {
	if len(t.Elements) == 0 {
		return Primitive{NIL}
	}
	if len(t.Elements) == 1 {
		return t.Elements[0].Type(ctx)
	}
	types := make([]ExpressionType, len(t.Elements))
	for i, element := range t.Elements {
		types[i] = element.Type(ctx)
	}
	return Tuple{types}
}

func (l TupleExpression) Check(c *Checker) {
	for _, el := range l.Elements {
		if el != nil {
			el.Check(c)
		}
	}

}

func (t TupleExpression) Loc() tokenizer.Loc {
	return t.loc
}

func ParseTupleExpression(p *Parser) Expression {
	lparen := p.tokenizer.Consume()
	loc := tokenizer.Loc{}
	loc.Start = lparen.Loc().Start

	elements := []Expression{}
	ParseList(p, tokenizer.RPAREN, func() {
		elements = append(elements, ParseTypedExpression(p))
	})

	next := p.tokenizer.Peek()
	if next.Kind() != tokenizer.RPAREN {
		p.report("')' expected", next.Loc())
	}
	loc.End = p.tokenizer.Consume().Loc().End

	return TupleExpression{elements, loc}
}
