package parser

import (
	"github.com/bmelicque/test-parser/tokenizer"
)

type IfElse struct {
	Keyword   tokenizer.Token
	Condition Node
	Alternate Node // IfElse | Body
	Body      *Block
}

func (i IfElse) Loc() tokenizer.Loc {
	return tokenizer.Loc{
		Start: i.Keyword.Loc().Start,
		End:   i.Body.Loc().End,
	}
}

func (p *Parser) parseIf() Node {
	keyword := p.tokenizer.Consume()
	outer := p.allowBraceParsing
	p.allowBraceParsing = false
	condition := ParseExpression(p)
	p.allowBraceParsing = outer
	body := p.parseBlock()
	alternate := parseAlternate(p)
	return IfElse{keyword, condition, alternate, body}
}

func parseAlternate(p *Parser) Node {
	if p.tokenizer.Peek().Kind() != tokenizer.ELSE_KW {
		return nil
	}
	p.tokenizer.Consume() // "else"
	switch p.tokenizer.Peek().Kind() {
	case tokenizer.IF_KW:
		return p.parseIf()
	case tokenizer.LBRACE:
		return *p.parseBlock()
	default:
		p.report("Block expected", p.tokenizer.Peek().Loc())
		return nil
	}
}
