package parser

import (
	"github.com/bmelicque/test-parser/tokenizer"
)

type TokenExpression struct {
	tokenizer.Token
}

func (t TokenExpression) Loc() tokenizer.Loc { return t.Token.Loc() }

func (TokenExpression) Parse(p *Parser) Node {
	token := p.tokenizer.Consume()
	switch token.Kind() {
	case tokenizer.BOOLEAN, tokenizer.NUMBER, tokenizer.STRING, tokenizer.IDENTIFIER, tokenizer.BOOL_KW, tokenizer.NUM_KW, tokenizer.STR_KW:
		return TokenExpression{token}
	}
	p.report("Expression expected", token.Loc())
	return nil
}
