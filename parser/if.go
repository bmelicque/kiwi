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

func ParseIf(p *Parser) Node {
	keyword := p.tokenizer.Consume()
	condition := ParseExpression(p)
	body := ParseBody(p)
	return IfElse{keyword, condition, body}
}
