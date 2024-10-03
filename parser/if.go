package parser

import (
	"github.com/bmelicque/test-parser/tokenizer"
)

type IfElse struct {
	Keyword   tokenizer.Token
	Condition Node
	Alternate Node // IfElse | Body
	Body      *Body
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
	body := p.parseBody()
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
		return *p.parseBody()
	default:
		p.report("Block expected", p.tokenizer.Peek().Loc())
		return nil
	}
}
