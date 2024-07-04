package parser

import (
	"github.com/bmelicque/test-parser/tokenizer"
)

type TupleExpression struct {
	elements []Expression
	loc      tokenizer.Loc
}

func (t TupleExpression) Type(ctx *Scope) ExpressionType {
	if len(t.elements) == 0 {
		return Primitive{NIL}
	}
	types := make([]ExpressionType, len(t.elements))
	for i, element := range t.elements {
		types[i] = element.Type(ctx)
	}
	return Tuple{types}
}

func (l TupleExpression) Check(c *Checker) {
	for _, el := range l.elements {
		if el != nil {
			el.Check(c)
		}
	}

}

func (t TupleExpression) Emit(e *Emitter) {
	e.Write("[")
	length := len(t.elements)
	for i, el := range t.elements {
		el.Emit(e)
		if i != length-1 {
			e.Write(", ")
		}
	}
	e.Write("]")
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
