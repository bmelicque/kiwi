package parser

import (
	"github.com/bmelicque/test-parser/tokenizer"
)

type TupleExpression struct {
	Elements []Node
}

func (t TupleExpression) Loc() tokenizer.Loc {
	return tokenizer.Loc{
		Start: t.Elements[0].Loc().Start,
		End:   t.Elements[len(t.Elements)-1].Loc().End,
	}
}

func (p *Parser) parseTupleExpression() Node {
	var elements []Node
	ParseList(p, tokenizer.ILLEGAL, func() {
		elements = append(elements, p.parseTypedExpression())
	})

	if len(elements) == 0 {
		return nil
	}
	if len(elements) == 1 {
		return elements[0]
	}
	return TupleExpression{elements}
}
