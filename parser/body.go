package parser

import (
	"github.com/bmelicque/test-parser/tokenizer"
)

type Body struct {
	Statements []Node
	loc        tokenizer.Loc
}

func (b Body) Loc() tokenizer.Loc { return b.loc }

func ParseBody(p *Parser) *Body {
	body := Body{}

	token := p.tokenizer.Consume()
	body.loc.Start = token.Loc().Start
	if token.Kind() != tokenizer.LBRACE {
		p.report("'{' expected", token.Loc())
	}

	body.Statements = []Node{}
	for p.tokenizer.Peek().Kind() != tokenizer.RBRACE && p.tokenizer.Peek().Kind() != tokenizer.EOF {
		body.Statements = append(body.Statements, ParseStatement(p))
	}

	token = p.tokenizer.Consume()
	body.loc.End = token.Loc().End
	if token.Kind() != tokenizer.RBRACE {
		p.report("'}' expected", token.Loc())
	}

	return &body
}
