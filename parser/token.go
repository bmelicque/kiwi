package parser

import (
	"github.com/bmelicque/test-parser/tokenizer"
)

type TokenExpression struct {
	tokenizer.Token
}

func (t TokenExpression) Loc() tokenizer.Loc { return t.Token.Loc() }

func (p *Parser) parseTokenExpression() Node {
	token := p.tokenizer.Peek()
	switch token.Kind() {
	case tokenizer.BOOLEAN, tokenizer.NUMBER, tokenizer.STRING, tokenizer.IDENTIFIER, tokenizer.BOOL_KW, tokenizer.NUM_KW, tokenizer.STR_KW:
		p.tokenizer.Consume()
		return TokenExpression{token}
	}
	if !p.allowEmptyExpr {
		p.tokenizer.Consume()
		p.report("Expression expected", token.Loc())
	}
	return nil
}
