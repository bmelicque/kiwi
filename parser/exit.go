package parser

import (
	"github.com/bmelicque/test-parser/tokenizer"
)

type Exit struct {
	Operator tokenizer.Token
	Value    Node
}

func (r Exit) Loc() tokenizer.Loc {
	loc := r.Operator.Loc()
	if r.Value != nil {
		loc.End = r.Value.Loc().End
	}
	return loc
}

func (p *Parser) parseExit() Exit {
	keyword := p.tokenizer.Consume()

	if p.tokenizer.Peek().Kind() == tokenizer.EOL {
		return Exit{keyword, nil}
	}

	value := ParseExpression(p)
	return Exit{keyword, value}
}
