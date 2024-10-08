package parser

type Block struct {
	Statements []Node
	loc        Loc
}

func (b Block) Loc() Loc { return b.loc }

func (p *Parser) parseBlock() *Block {
	block := Block{}

	token := p.Consume()
	block.loc.Start = token.Loc().Start
	if token.Kind() != LBRACE {
		p.report("'{' expected", token.Loc())
	}
	p.DiscardLineBreaks()

	block.Statements = []Node{}
	for p.Peek().Kind() != RBRACE && p.Peek().Kind() != EOF {
		block.Statements = append(block.Statements, p.parseStatement())
		p.DiscardLineBreaks()
	}

	token = p.Consume()
	block.loc.End = token.Loc().End
	if token.Kind() != RBRACE {
		p.report("'}' expected", token.Loc())
	}

	return &block
}
