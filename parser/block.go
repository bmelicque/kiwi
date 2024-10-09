package parser

type Block struct {
	Statements []Node
	loc        Loc
}

func (b Block) Loc() Loc { return b.loc }
func (b Block) reportLoc() Loc {
	if len(b.Statements) > 0 {
		return b.Statements[len(b.Statements)-1].Loc()
	} else {
		return b.loc
	}
}
func (b Block) Type() ExpressionType {
	if len(b.Statements) == 0 {
		return Primitive{NIL}
	}
	last := b.Statements[len(b.Statements)-1]
	expr, ok := last.(Expression)
	if !ok {
		return Primitive{NIL}
	}
	return expr.Type()
}

func (p *Parser) parseBlock() *Block {
	block := Block{}

	token := p.Consume()
	block.loc.Start = token.Loc().Start
	if token.Kind() != LeftBrace {
		p.report("'{' expected", token.Loc())
	}
	p.DiscardLineBreaks()

	block.Statements = []Node{}
	for p.Peek().Kind() != RightBrace && p.Peek().Kind() != EOF {
		block.Statements = append(block.Statements, p.parseStatement())
		p.DiscardLineBreaks()
	}

	token = p.Consume()
	block.loc.End = token.Loc().End
	if token.Kind() != RightBrace {
		p.report("'}' expected", token.Loc())
	}

	reportUnreachableCode(p, block.Statements)

	return &block
}

func reportUnreachableCode(p *Parser, statements []Node) {
	var foundExit bool
	var unreachable Loc
	for _, statement := range statements {
		if foundExit {
			if unreachable.Start == (Position{}) {
				unreachable.Start = statement.Loc().Start
			}
			unreachable.End = statement.Loc().End
		}
		if _, ok := statement.(Exit); ok {
			foundExit = true
		}
	}
	if unreachable != (Loc{}) {
		p.report("Detected unreachable code", unreachable)
	}
}
