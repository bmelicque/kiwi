package parser

import "slices"

type Block struct {
	Statements []Node
	loc        Loc
}

func (b *Block) typeCheck(p *Parser) {
	for i := range b.Statements {
		b.Statements[i].typeCheck(p)
	}
	if b.Type().Kind() == TYPE {
		p.report("Blocks shouldn't return types", p.parseBlock().reportLoc())
	}
}

func (b *Block) Loc() Loc { return b.loc }
func (b *Block) reportLoc() Loc {
	if len(b.Statements) > 0 {
		return b.Statements[len(b.Statements)-1].Loc()
	} else {
		return b.loc
	}
}
func (b *Block) Type() ExpressionType {
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
	if p.Peek().Kind() != LeftBrace {
		p.report("'{' expected", p.Peek().Loc())
		return nil
	}

	start := p.Consume().Loc().Start // '{'
	p.DiscardLineBreaks()

	statements := []Node{}
	stopAt := []TokenKind{RightBrace, EOL, EOF}
	for p.Peek().Kind() != RightBrace && p.Peek().Kind() != EOF {
		statements = append(statements, p.parseStatement())
		if !slices.Contains(stopAt, p.Peek().Kind()) {
			recover(p, RightBrace)
		}
		p.DiscardLineBreaks()
	}
	reportUnreachableCode(p, statements)

	if p.Peek().Kind() != RightBrace {
		p.report("'}' expected", p.Peek().Loc())
	}
	end := p.Consume().Loc().End

	return &Block{statements, Loc{start, end}}
}

func reportUnreachableCode(p *Parser, statements []Node) {
	var foundExit, foundUnreachable bool
	var unreachable Loc
	for _, statement := range statements {
		if foundExit {
			foundUnreachable = true
			if unreachable.Start == (Position{}) {
				unreachable.Start = statement.Loc().Start
			}
			unreachable.End = statement.Loc().End
		}
		if _, ok := statement.(*Exit); ok {
			foundExit = true
		}
	}
	if foundUnreachable {
		p.report("Detected unreachable code", unreachable)
	}
}
