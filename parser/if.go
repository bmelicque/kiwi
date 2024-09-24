package parser

import (
	"github.com/bmelicque/test-parser/tokenizer"
)

type IfElse struct {
	Keyword   tokenizer.Token
	Condition Node
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
	return IfElse{keyword, condition, body}
}
