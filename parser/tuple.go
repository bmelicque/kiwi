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
	// TODO: remove parseList...

	// parseList(p, tokenizer.ILLEGAL, func() {
	// 	elements = append(elements, p.parseSumType())
	// })

	outer := p.allowEmptyExpr
	p.allowEmptyExpr = true
	for p.tokenizer.Peek().Kind() != tokenizer.EOF {
		el := p.parseSumType()
		if el == nil {
			break
		}
		elements = append(elements, el)

		if p.tokenizer.Peek().Kind() != tokenizer.COMMA {
			break
		}
		p.tokenizer.Consume()

		if p.multiline {
			p.tokenizer.DiscardLineBreaks()
		}
	}
	p.allowEmptyExpr = outer

	if len(elements) == 0 {
		return nil
	}
	if len(elements) == 1 {
		return elements[0]
	}
	return TupleExpression{elements}
}
