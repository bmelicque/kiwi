package parser

import (
	"github.com/bmelicque/test-parser/tokenizer"
)

type Return struct {
	Operator tokenizer.Token
	Value    Node
}

func (r Return) Loc() tokenizer.Loc {
	loc := r.Operator.Loc()
	if r.Value != nil {
		loc.End = r.Value.Loc().End
	}
	return loc
}

func ParseReturn(p *Parser) Node {
	keyword := p.tokenizer.Consume()

	if p.tokenizer.Peek().Kind() == tokenizer.EOL {
		return Return{keyword, nil}
	}

	value := ParseExpression(p)
	return Return{keyword, value}
}
