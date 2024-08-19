package parser

import (
	"github.com/bmelicque/test-parser/tokenizer"
)

type TupleExpression struct {
	Elements []Node
	loc      tokenizer.Loc
}

func (t TupleExpression) Loc() tokenizer.Loc {
	return t.loc
}

func parseTupleExpression(p *Parser) Node {
	lparen := p.tokenizer.Consume()
	loc := tokenizer.Loc{}
	loc.Start = lparen.Loc().Start

	var elements []Node
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
