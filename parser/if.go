package parser

type IfElse struct {
	Keyword   Token
	Condition Node
	Alternate Node // IfElse | Body
	Body      *Block
}

func (i IfElse) Loc() Loc {
	return Loc{
		Start: i.Keyword.Loc().Start,
		End:   i.Body.Loc().End,
	}
}

func (p *Parser) parseIf() Node {
	keyword := p.Consume()
	outer := p.allowBraceParsing
	p.allowBraceParsing = false
	condition := ParseExpression(p)
	p.allowBraceParsing = outer
	body := p.parseBlock()
	alternate := parseAlternate(p)
	return IfElse{keyword, condition, alternate, body}
}

func parseAlternate(p *Parser) Node {
	if p.Peek().Kind() != ELSE_KW {
		return nil
	}
	p.Consume() // "else"
	switch p.Peek().Kind() {
	case IF_KW:
		return p.parseIf()
	case LBRACE:
		return *p.parseBlock()
	default:
		p.report("Block expected", p.Peek().Loc())
		return nil
	}
}
