package parser

import (
	"github.com/bmelicque/test-parser/tokenizer"
)

type Block struct {
	Statements []Node
	loc        tokenizer.Loc
}

func (b Block) Loc() tokenizer.Loc { return b.loc }

func (p *Parser) parseBlock() *Block {
	block := Block{}

	token := p.tokenizer.Consume()
	block.loc.Start = token.Loc().Start
	if token.Kind() != tokenizer.LBRACE {
		p.report("'{' expected", token.Loc())
	}
	p.tokenizer.DiscardLineBreaks()

	block.Statements = []Node{}
	for p.tokenizer.Peek().Kind() != tokenizer.RBRACE && p.tokenizer.Peek().Kind() != tokenizer.EOF {
		block.Statements = append(block.Statements, p.parseStatement())
		p.tokenizer.DiscardLineBreaks()
	}

	token = p.tokenizer.Consume()
	block.loc.End = token.Loc().End
	if token.Kind() != tokenizer.RBRACE {
		p.report("'}' expected", token.Loc())
	}

	return &block
}
