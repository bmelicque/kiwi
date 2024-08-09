package parser

import (
	"github.com/bmelicque/test-parser/tokenizer"
)

type ListExpression struct {
	Elements []Node
	loc      tokenizer.Loc
}

func (l ListExpression) Loc() tokenizer.Loc { return l.loc }
func (l ListExpression) Parse(p *Parser) Node {
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
